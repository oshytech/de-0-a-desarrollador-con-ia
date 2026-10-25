"""LLMOps mínimo para el asistente del club: trazas, coste, caché y evaluación.

Uso:
    python -I llmops.py preguntar "¿Cuánto dura una reserva?" [--version v2]
    python -I llmops.py evaluar [--versiones v1 v2] [--umbral 0.9] [--sin-cache]
    python -I llmops.py informe

Modo real: define LLM_BASE_URL (cualquier API compatible con OpenAI, también un
modelo local) y, si hace falta, LLM_API_KEY y LLM_MODEL. Sin ninguna de las
dos primeras, funciona en modo simulado: un "modelo" de mentira, determinista,
que sirve para probar todo lo demás sin gastar nada.
"""

import argparse
import hashlib
import json
import os
import re
import sqlite3
import statistics
import sys
import time
import urllib.request
import uuid
from contextlib import contextmanager
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8")

DIR = Path(__file__).resolve().parent
TRABAJO = DIR / ".llmops"
NORMAS = (DIR / "datos" / "normas.md").read_text(encoding="utf-8")

# Precios en dólares por millón de tokens (entrada, salida).
# VALORES DE EJEMPLO: consulta siempre la tarifa actual de tu proveedor.
PRECIOS = {
    "simulado-pequeno": (0.15, 0.60),
    "simulado-grande": (3.00, 15.00),
}

# Los prompts tienen versión, igual que el código. Cambiar uno es un cambio
# que se evalúa antes de llegar a producción.
PROMPTS = {
    "v1": "Eres el asistente del club de tenis y pádel. Responde en una o dos frases.",
    "v2": (
        "Eres el asistente del club de tenis y pádel. Responde en una o dos frases "
        "usando solo las normas de abajo. Si la respuesta no está en ellas, di "
        "exactamente: No tengo esa información.\n\nNORMAS DEL CLUB:\n" + NORMAS
    ),
}

PROMPT_JUEZ = (
    "Eres un evaluador estricto. Recibes una PREGUNTA, una REFERENCIA correcta y "
    "una RESPUESTA. Contesta solo CORRECTA si la respuesta dice lo mismo que la "
    "referencia, aunque use otras palabras, o INCORRECTA si no."
)


def coste(modelo, entrada, salida):
    """Coste en dólares de una llamada, o None si no conocemos el precio."""
    if modelo not in PRECIOS:
        return None
    precio_entrada, precio_salida = PRECIOS[modelo]
    return (entrada * precio_entrada + salida * precio_salida) / 1_000_000


# --- Datos personales ---------------------------------------------------------

CORREO = re.compile(r"[\w.+-]+@[\w-]+\.[\w.]+")
TELEFONO = re.compile(r"\b\d{3} ?\d{3} ?\d{3}\b")


def redactar(texto):
    """Quita correos y teléfonos antes de enviar o guardar el texto."""
    return TELEFONO.sub("[teléfono]", CORREO.sub("[correo]", texto))


# --- Trazas -------------------------------------------------------------------

class Trazador:
    """Escribe un span por línea en un fichero JSONL. Cada span sabe de qué
    traza forma parte y quién es su padre, así se reconstruye el árbol."""

    def __init__(self, fichero):
        self.fichero = fichero
        self.pila = []
        self.coste_acumulado = 0.0

    @contextmanager
    def span(self, nombre, **atributos):
        padre = self.pila[-1] if self.pila else None
        registro = {
            "traza": padre["traza"] if padre else uuid.uuid4().hex[:12],
            "span": uuid.uuid4().hex[:8],
            "padre": padre["span"] if padre else None,
            "nombre": nombre,
            "inicio": time.strftime("%Y-%m-%dT%H:%M:%S"),
            **atributos,
        }
        self.pila.append(registro)
        t0 = time.perf_counter()
        try:
            yield registro
        except Exception as e:
            registro["error"] = repr(e)
            raise
        finally:
            registro["ms"] = round((time.perf_counter() - t0) * 1000, 1)
            self.coste_acumulado += registro.get("coste_usd") or 0
            self.pila.pop()
            with self.fichero.open("a", encoding="utf-8") as f:
                f.write(json.dumps(registro, ensure_ascii=False) + "\n")


# --- Caché --------------------------------------------------------------------

def clave_cache(modelo, version, temperatura, mensajes):
    """Todo lo que cambia la respuesta forma parte de la clave."""
    material = json.dumps([modelo, version, temperatura, mensajes], ensure_ascii=False)
    return hashlib.sha256(material.encode("utf-8")).hexdigest()


class Cache:
    def __init__(self, fichero, ttl_segundos=24 * 3600):
        self.ttl = ttl_segundos
        self.bd = sqlite3.connect(fichero)
        self.bd.execute(
            "CREATE TABLE IF NOT EXISTS cache (clave TEXT PRIMARY KEY, "
            "respuesta TEXT, guardada REAL)"
        )

    def leer(self, clave):
        fila = self.bd.execute(
            "SELECT respuesta, guardada FROM cache WHERE clave = ?", (clave,)
        ).fetchone()
        if fila is None or time.time() - fila[1] > self.ttl:
            return None
        return json.loads(fila[0])

    def guardar(self, clave, respuesta):
        self.bd.execute(
            "INSERT OR REPLACE INTO cache VALUES (?, ?, ?)",
            (clave, json.dumps(respuesta, ensure_ascii=False), time.time()),
        )
        self.bd.commit()


class SinCache:
    def leer(self, clave):
        return None

    def guardar(self, clave, respuesta):
        pass


# --- Cliente del modelo -------------------------------------------------------

def estimar_tokens(texto):
    # Aproximación habitual: unos 4 caracteres por token.
    return len(texto) // 4 + 1


class Cliente:
    def __init__(self):
        self.base_url = os.environ.get("LLM_BASE_URL", "").rstrip("/")
        self.clave = os.environ.get("LLM_API_KEY", "")
        self.simulado = not (self.base_url or self.clave)
        if self.simulado:
            self.modelo = "simulado-pequeno"
        else:
            self.base_url = self.base_url or "https://api.openai.com/v1"
            self.modelo = os.environ.get("LLM_MODEL") or sys.exit("Falta LLM_MODEL")

    def completar(self, mensajes, temperatura=0.0):
        """Devuelve {"texto", "entrada", "salida"} (tokens de entrada y salida)."""
        if self.simulado:
            return self._simular(mensajes)
        cuerpo = json.dumps(
            {"model": self.modelo, "messages": mensajes, "temperature": temperatura}
        ).encode("utf-8")
        peticion = urllib.request.Request(
            self.base_url + "/chat/completions",
            data=cuerpo,
            headers={"Content-Type": "application/json"}
            | ({"Authorization": f"Bearer {self.clave}"} if self.clave else {}),
        )
        with urllib.request.urlopen(peticion, timeout=60) as r:
            datos = json.load(r)
        texto = datos["choices"][0]["message"]["content"].strip()
        uso = datos.get("usage") or {}
        return {
            "texto": texto,
            "entrada": uso.get("prompt_tokens") or estimar_tokens(json.dumps(mensajes)),
            "salida": uso.get("completion_tokens") or estimar_tokens(texto),
        }

    # Modo simulado. No es un modelo: es un sustituto determinista que se
    # comporta como uno bueno si el prompt trae las normas y como uno que
    # se lo inventa si no las trae.
    def _simular(self, mensajes):
        sistema, usuario = mensajes[0]["content"], mensajes[-1]["content"]
        if sistema.startswith("Eres un evaluador"):
            texto = _juez_simulado(usuario)
        elif "NORMAS DEL CLUB:" in sistema:
            texto = _norma_mas_parecida(sistema, usuario)
        else:
            texto = "Las reservas duran 90 minutos y puedes cancelar hasta 24 horas antes."
        time.sleep(0.02 + len(sistema) / 200_000)  # latencia de mentira
        return {
            "texto": texto,
            "entrada": estimar_tokens(sistema + usuario),
            "salida": estimar_tokens(texto),
        }


def _palabras(texto):
    """Palabras de 4 o más letras y números (las horas cuentan doble)."""
    peso = {}
    for p in re.findall(r"\d+:\d+|\w+", texto.lower()):
        if any(c.isdigit() for c in p):
            peso[p] = 2
        elif len(p) >= 4:
            peso[p] = 1
    return peso


def _norma_mas_parecida(sistema, pregunta):
    normas = [l[2:].strip() for l in sistema.splitlines() if l.startswith("- ")]
    buscadas = _palabras(pregunta)
    puntos = [sum(buscadas.get(p, 0) for p in _palabras(n)) for n in normas]
    mejor = max(range(len(normas)), key=lambda i: puntos[i])
    return normas[mejor] if puntos[mejor] > 0 else "No tengo esa información."


def _juez_simulado(contenido):
    campos = dict(re.findall(r"^(PREGUNTA|REFERENCIA|RESPUESTA): (.*)$", contenido, re.M))
    referencia = _palabras(campos.get("REFERENCIA", ""))
    respuesta = _palabras(campos.get("RESPUESTA", ""))
    numeros = [p for p, peso in referencia.items() if peso == 2]
    largas = [p for p, peso in referencia.items() if peso == 1 and len(p) >= 5]
    numeros_ok = all(n in respuesta for n in numeros)
    parecido = sum(p in respuesta for p in largas) / max(len(largas), 1)
    return "CORRECTA" if numeros_ok and parecido >= 0.6 else "INCORRECTA"


# --- El asistente -------------------------------------------------------------

def responder(pregunta, version, cliente, cache, trazador, temperatura=0.0):
    pregunta = redactar(pregunta)
    mensajes = [
        {"role": "system", "content": PROMPTS[version]},
        {"role": "user", "content": pregunta},
    ]
    with trazador.span("responder", version_prompt=version, pregunta=pregunta) as span:
        # Solo se cachea lo que pedimos determinista (temperatura 0).
        clave = clave_cache(cliente.modelo, version, temperatura, mensajes)
        guardada = cache.leer(clave) if temperatura == 0 else None
        if guardada:
            span["cache"] = "acierto"
            return guardada["texto"]
        span["cache"] = "fallo"
        with trazador.span("llm", modelo=cliente.modelo) as llm:
            r = cliente.completar(mensajes, temperatura)
            llm["tokens_entrada"], llm["tokens_salida"] = r["entrada"], r["salida"]
            llm["coste_usd"] = coste(cliente.modelo, r["entrada"], r["salida"])
        cache.guardar(clave, r)
        return r["texto"]


def juzgar(ejemplo, respuesta, cliente, trazador):
    contenido = (
        f"PREGUNTA: {redactar(ejemplo['pregunta'])}\n"
        f"REFERENCIA: {ejemplo['referencia']}\n"
        f"RESPUESTA: {respuesta}"
    )
    mensajes = [
        {"role": "system", "content": PROMPT_JUEZ},
        {"role": "user", "content": contenido},
    ]
    with trazador.span("juez", modelo=cliente.modelo) as span:
        r = cliente.completar(mensajes, 0.0)
        span["tokens_entrada"], span["tokens_salida"] = r["entrada"], r["salida"]
        span["coste_usd"] = coste(cliente.modelo, r["entrada"], r["salida"])
        span["veredicto"] = r["texto"]
    return r["texto"].strip().upper().startswith("CORRECTA")


def contiene(ejemplo, respuesta):
    """Evaluador determinista: cada grupo exige al menos una de sus variantes."""
    texto = respuesta.lower()
    return all(any(v.lower() in texto for v in grupo) for grupo in ejemplo["debe_contener"])


# --- Comandos -----------------------------------------------------------------

def preparar(sin_cache=False):
    TRABAJO.mkdir(exist_ok=True)
    trazador = Trazador(TRABAJO / "trazas.jsonl")
    cache = SinCache() if sin_cache else Cache(TRABAJO / "cache.sqlite")
    return Cliente(), cache, trazador


def cmd_preguntar(args):
    cliente, cache, trazador = preparar(args.sin_cache)
    print(responder(args.pregunta, args.version, cliente, cache, trazador))


def cmd_evaluar(args):
    cliente, cache, trazador = preparar(args.sin_cache)
    dorado = [
        json.loads(l)
        for l in (DIR / "datos" / "dorado.jsonl").read_text(encoding="utf-8").splitlines()
        if l.strip()
    ]
    modo = "simulado" if cliente.simulado else "real"
    print(f"Modelo: {cliente.modelo} ({modo}). Ejemplos: {len(dorado)}\n")
    print(f"{'versión':<8}{'contiene':>10}{'juez':>8}{'p50 ms':>9}{'coste USD':>12}  fallos")
    resultado = {}
    for version in args.versiones:
        fallos, ok_contiene, ok_juez, tiempos = [], 0, 0, []
        gasto_previo = trazador.coste_acumulado
        for ejemplo in dorado:
            t0 = time.perf_counter()
            respuesta = responder(ejemplo["pregunta"], version, cliente, cache, trazador)
            tiempos.append((time.perf_counter() - t0) * 1000)
            a = contiene(ejemplo, respuesta)
            b = juzgar(ejemplo, respuesta, cliente, trazador)
            ok_contiene += a
            ok_juez += b
            if not (a and b):
                fallos.append(ejemplo["id"])
        gasto = trazador.coste_acumulado - gasto_previo
        n = len(dorado)
        resultado[version] = ok_juez / n
        print(
            f"{version:<8}{ok_contiene / n:>10.0%}{ok_juez / n:>8.0%}"
            f"{statistics.median(tiempos):>9.0f}{gasto:>12.6f}  {', '.join(fallos) or '-'}"
        )
    ultima = args.versiones[-1]
    if resultado[ultima] < args.umbral:
        print(f"\nLa versión {ultima} no llega al umbral ({args.umbral:.0%}).")
        sys.exit(1)


def cmd_informe(args):
    fichero = TRABAJO / "trazas.jsonl"
    if not fichero.exists():
        sys.exit("Todavía no hay trazas. Ejecuta antes 'preguntar' o 'evaluar'.")
    spans = [json.loads(l) for l in fichero.read_text(encoding="utf-8").splitlines()]
    respuestas = [s for s in spans if s["nombre"] == "responder"]
    llamadas = [s for s in spans if s["nombre"] in ("llm", "juez")]
    ms = sorted(s["ms"] for s in respuestas)
    aciertos = sum(s.get("cache") == "acierto" for s in respuestas)
    print(f"Respuestas del asistente: {len(respuestas)}")
    print(f"Aciertos de caché:        {aciertos} ({aciertos / len(respuestas):.0%})")
    print(f"Latencia p50 / p95:       {ms[len(ms) // 2]:.0f} / {ms[int(len(ms) * 0.95)]:.0f} ms")
    print(f"Llamadas al modelo:       {len(llamadas)}")
    print(f"Tokens entrada / salida:  {sum(s.get('tokens_entrada', 0) for s in llamadas)}"
          f" / {sum(s.get('tokens_salida', 0) for s in llamadas)}")
    print(f"Coste total:              {sum((s.get('coste_usd') or 0) for s in llamadas):.6f} USD")
    errores = [s for s in spans if "error" in s]
    print(f"Spans con error:          {len(errores)}")


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    sub = p.add_subparsers(required=True)

    a = sub.add_parser("preguntar")
    a.add_argument("pregunta")
    a.add_argument("--version", default="v2", choices=PROMPTS)
    a.add_argument("--sin-cache", action="store_true")
    a.set_defaults(func=cmd_preguntar)

    e = sub.add_parser("evaluar")
    e.add_argument("--versiones", nargs="+", default=["v1", "v2"], choices=PROMPTS)
    e.add_argument("--umbral", type=float, default=0.9)
    e.add_argument("--sin-cache", action="store_true")
    e.set_defaults(func=cmd_evaluar)

    i = sub.add_parser("informe")
    i.set_defaults(func=cmd_informe)

    args = p.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
