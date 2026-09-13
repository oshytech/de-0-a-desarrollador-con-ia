"""Cuatro patrones de orquestación aplicados a los mensajes de los socios del club.

  1. Enrutar: un clasificador barato decide qué especialista atiende el mensaje.
  2. Encadenar con puertas: cada paso del modelo pasa por una comprobación en código.
  3. Paralelizar: tres revisores independientes y un agregador determinista.
  4. Generar y evaluar: un bucle con criterios explícitos y un presupuesto de vueltas.

cliente_ia.py es una copia del capítulo 27. Sin LLM_API_KEY ni LLM_BASE_URL todo
funciona en modo simulado: las funciones simular_* imitan las respuestas de un
modelo para que el flujo se pueda ejecutar y probar sin coste.

Uso: python -I orquestacion.py
"""
import json
import re
import sys
import time
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))  # para importar cliente_ia también con python -I
from cliente_ia import preguntar

sys.stdout.reconfigure(encoding="utf-8")

PISTAS = {"P1": ("Tenis 1", "tenis"), "P2": ("Pádel 1", "padel"), "P3": ("Pádel 2", "padel")}
OCUPADAS = {("P2", "2026-12-03T19:00"), ("P1", "2026-12-03T09:00")}


class ErrorDePuerta(Exception):
    def __init__(self, puerta, motivo):
        super().__init__(f"Puerta {puerta}: {motivo}")
        self.puerta = puerta


class Escalar(Exception):
    """El bucle agotó su presupuesto: decide una persona."""


# ================================================================ 1. Enrutar

CATEGORIAS = {"RESERVA", "CANCELACION", "INCIDENCIA"}
CLASIFICADOR = ("Clasifica el mensaje de un socio de un club deportivo. Responde con UNA sola palabra: "
                "RESERVA, CANCELACION o INCIDENCIA. Si no encaja en ninguna, responde OTRO.")


def simular_clasificador(_sistema, usuario):
    texto = usuario.lower()
    if re.search(r"anula|cancel|no puedo ir", texto):
        return "CANCELACION"
    if re.search(r"rota|roto|no funciona|sucia|luz", texto):
        return "INCIDENCIA"
    if re.search(r"reserv|pista|jugar", texto):
        return "RESERVA"
    return "OTRO"


def enrutar(mensaje, simulador=simular_clasificador):
    """Devuelve la categoría o FUERA si la respuesta no está en la taxonomía."""
    palabra = preguntar(CLASIFICADOR, mensaje, simulador, max_tokens=5).texto.strip().upper()
    return palabra if palabra in CATEGORIAS else "FUERA"


# ===================================================== 2. Encadenar con puertas

EXTRACTOR = ('Extrae de este mensaje la reserva pedida. Responde SOLO con JSON: '
             '{"deporte": "tenis" o "padel", "inicio": "AAAA-MM-DDTHH:00"}. Hoy es martes 2026-12-01.')
REDACTOR = ("Escribe una respuesta breve y amable al socio confirmando la pista, el día y la hora "
            "que te doy. No prometas nada más.")


def simular_extractor(_sistema, usuario):
    deporte = "padel" if "pádel" in usuario.lower() else "tenis"
    hora = re.search(r"a las (\d{1,2})", usuario)
    h = int(hora.group(1)) + (12 if "tarde" in usuario and int(hora.group(1)) < 12 else 0) if hora else 18
    return json.dumps({"deporte": deporte, "inicio": f"2026-12-03T{h:02d}:00"})


def simular_redactor(_sistema, usuario):
    datos = json.loads(usuario)
    return (f"Hola. Te hemos reservado {datos['pista']} el {datos['dia']} a las {datos['hora']}. "
            "Si no puedes venir, cancela con al menos dos horas de antelación.")


def cadena_reserva(mensaje, extractor=simular_extractor, redactor=simular_redactor):
    # Paso 1 (modelo): extraer los datos del mensaje.
    texto = preguntar(EXTRACTOR, mensaje, extractor, max_tokens=60).texto
    # Puerta 1 (código): el JSON tiene la forma acordada y una hora del horario del club.
    try:
        peticion = json.loads(texto)
        inicio = datetime.fromisoformat(peticion["inicio"])
    except (json.JSONDecodeError, KeyError, TypeError, ValueError):
        raise ErrorDePuerta(1, "la extracción no tiene la forma acordada") from None
    if peticion.get("deporte") not in {"tenis", "padel"} or inicio.minute or not 8 <= inicio.hour <= 21:
        raise ErrorDePuerta(1, f"datos fuera del esquema o del horario: {peticion}")

    # Paso 2 (código, no modelo): buscar una pista libre es determinista.
    clave = inicio.strftime("%Y-%m-%dT%H:%M")
    libres = [pid for pid, (_, deporte) in PISTAS.items()
              if deporte == peticion["deporte"] and (pid, clave) not in OCUPADAS]
    # Puerta 2 (código): hay algo que ofrecer.
    if not libres:
        raise ErrorDePuerta(2, "no hay pistas libres a esa hora")
    nombre = PISTAS[libres[0]][0]
    dia, hora = clave.split("T")

    # Paso 3 (modelo): redactar la respuesta.
    datos = json.dumps({"pista": nombre, "dia": dia, "hora": hora}, ensure_ascii=False)
    respuesta = preguntar(REDACTOR, datos, redactor, max_tokens=120).texto
    # Puerta 3 (código): la respuesta dice lo que se ha reservado de verdad.
    if not all(dato in respuesta for dato in (nombre, dia, hora)):
        raise ErrorDePuerta(3, "la respuesta no menciona la pista, el día y la hora reales")
    return respuesta


# ================================================ 3. Paralelizar y agregar

CODIGO_A_REVISAR = '''\
def puede_cancelar(reserva, ahora):
    return reserva["inicio"] - ahora > timedelta(hours=2)
'''
DIMENSIONES = {
    "seguridad": "Revisa SOLO la seguridad: quién puede hacer qué y datos sin validar.",
    "robustez": "Revisa SOLO la robustez: casos límite, estados y errores sin tratar.",
    "legibilidad": "Revisa SOLO la legibilidad: nombres, documentación y claridad.",
}
FORMATO = (' Responde SOLO con una lista JSON de hallazgos: '
           '[{"linea": n, "severidad": "alta|media|baja", "hallazgo": "..."}]. Lista vacía si no hay.')
HALLAZGOS_SIMULADOS = {
    "seguridad": [{"linea": 1, "severidad": "alta",
                   "hallazgo": "No comprueba que la reserva sea del socio que cancela."}],
    "robustez": [{"linea": 2, "severidad": "alta",
                  "hallazgo": "Con exactamente 2 horas de margen no deja cancelar: debería ser >=."},
                 {"linea": 2, "severidad": "media", "hallazgo": "No comprueba que la reserva esté confirmada."}],
    "legibilidad": [{"linea": 2, "severidad": "media", "hallazgo": "No comprueba que la reserva esté confirmada."},
                    {"linea": 1, "severidad": "baja", "hallazgo": "Falta explicar la regla de las 2 horas."}],
}
DEMORA_SIMULADA = {"seguridad": 0.2, "robustez": 0.3, "legibilidad": 0.4}


def simular_revisor(sistema, _usuario):
    dimension = next(d for d in DIMENSIONES if DIMENSIONES[d] in sistema)
    time.sleep(DEMORA_SIMULADA[dimension])  # cada revisor tarda lo suyo
    return json.dumps(HALLAZGOS_SIMULADOS[dimension], ensure_ascii=False)


def revisar(dimension, codigo, simulador):
    texto = preguntar(DIMENSIONES[dimension] + FORMATO, codigo, simulador, max_tokens=400).texto
    try:
        hallazgos = json.loads(texto)
    except json.JSONDecodeError:
        return dimension, None  # el agregador lo cuenta como revisor fallido
    return dimension, hallazgos


def revisar_en_paralelo(codigo, simulador=simular_revisor):
    with ThreadPoolExecutor(max_workers=len(DIMENSIONES)) as grupo:
        resultados = list(grupo.map(lambda d: revisar(d, codigo, simulador), DIMENSIONES))
    # Agregador determinista: valida, quita duplicados y ordena por severidad.
    orden = {"alta": 0, "media": 1, "baja": 2}
    vistos, informe, fallidos = set(), [], []
    for dimension, hallazgos in resultados:
        if hallazgos is None:
            fallidos.append(dimension)
            continue
        for h in hallazgos:
            if not isinstance(h, dict) or h.get("severidad") not in orden:
                continue
            clave = (h.get("linea"), h["hallazgo"].strip().lower())
            if clave not in vistos:
                vistos.add(clave)
                informe.append(dict(h, dimension=dimension))
    informe.sort(key=lambda h: (orden[h["severidad"]], h["linea"]))
    return informe, fallidos


# ================================================== 4. Generar y evaluar

def simular_generador(_sistema, usuario):
    datos = json.loads(usuario)
    texto = f"Hola, {datos['socio']}. Tu reserva de {datos['pista']} del {datos['dia']} queda confirmada."
    if any("hora" in f for f in datos["correcciones"]):
        texto = texto.replace(" queda confirmada.", f" a las {datos['hora']} queda confirmada.")
    return texto


def evaluar(texto, reserva):
    """Criterios explícitos, comprobados en código. Devuelve la lista de fallos."""
    fallos = []
    for campo, nombre in (("pista", "la pista"), ("dia", "el día"), ("hora", "la hora")):
        if reserva[campo] not in texto:
            fallos.append(f"Falta {nombre}: {reserva[campo]}.")
    if len(texto.split()) > 60:
        fallos.append("Tiene más de 60 palabras.")
    if re.search(r"garantiz|seguro que|sin falta", texto, re.I):
        fallos.append("Promete algo que el club no ha decidido.")
    return fallos


def generar_y_evaluar(reserva, generador=simular_generador, max_vueltas=3):
    correcciones = []
    for vuelta in range(1, max_vueltas + 1):
        peticion = json.dumps(dict(reserva, correcciones=correcciones), ensure_ascii=False)
        texto = preguntar("Escribe un aviso de reserva para el socio. Aplica las correcciones.",
                          peticion, generador, max_tokens=150).texto
        fallos = evaluar(texto, reserva)
        print(f"  vuelta {vuelta}: {'PASA' if not fallos else 'FALLA ' + ' '.join(fallos)}")
        if not fallos:
            return texto, vuelta
        correcciones = fallos  # solo vuelve lo que hay que corregir
    raise Escalar(f"{max_vueltas} vueltas sin pasar la evaluación: {correcciones}")


# ===================================================================== demo

if __name__ == "__main__":
    print("1. Enrutar")
    for mensaje in ["¿Hay pista de pádel el jueves a las 7 de la tarde?",
                    "No puedo ir mañana, anulad mi reserva",
                    "La red de la pista 2 está rota",
                    "¿A qué hora abre la cafetería?"]:
        print(f"  {enrutar(mensaje):<11} <- {mensaje}")

    print("\n2. Encadenar con puertas")
    for mensaje in ["Pádel el jueves a las 7 de la tarde", "Tenis el jueves a las 9"]:
        try:
            print(f"  OK: {cadena_reserva(mensaje)}")
        except ErrorDePuerta as e:
            print(f"  Cortado. {e}")

    print("\n3. Revisar en paralelo")
    inicio = time.monotonic()
    informe, fallidos = revisar_en_paralelo(CODIGO_A_REVISAR)
    print(f"  {len(informe)} hallazgos sin duplicados en {time.monotonic() - inicio:.1f} s "
          f"(en serie serían {sum(DEMORA_SIMULADA.values()):.1f} s)")
    for h in informe:
        print(f"  [{h['severidad']}] línea {h['linea']} ({h['dimension']}): {h['hallazgo']}")

    print("\n4. Generar y evaluar")
    reserva = {"socio": "Lucía", "pista": "Pádel 2", "dia": "2026-12-03", "hora": "19:00"}
    texto, vueltas = generar_y_evaluar(reserva)
    print(f"  Aprobado en la vuelta {vueltas}: {texto}")
