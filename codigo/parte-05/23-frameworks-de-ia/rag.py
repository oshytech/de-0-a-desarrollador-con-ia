"""Un RAG mínimo sobre la normativa del club, sin librerías externas.

Las piezas son las de cualquier RAG: trocear documentos, convertir cada
trozo en un vector, buscar los trozos más parecidos a la pregunta y pasárselos
al modelo con la instrucción de citar la fuente.

Los vectores de este ejemplo son TF-IDF sobre raíces de palabras, para que
funcione sin descargar nada. Un sistema real usa un modelo de embeddings,
que también encuentra sinónimos y frases parecidas con palabras distintas.

Uso: python rag.py "¿Cuántos invitados puedo llevar?"
"""
import math
import os
import re
import sys
import unicodedata
from collections import Counter
from dataclasses import dataclass
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))  # para importar cliente_llm también con python -I
from cliente_llm import completar

sys.stdout.reconfigure(encoding="utf-8")

CARPETA = Path(__file__).parent / "documentos"
VACIAS = set("a al de del el en la las lo los y o un una que se por para con es hay mi me".split())
UMBRAL = 0.25  # por debajo, el fragmento se considera irrelevante


@dataclass
class Fragmento:
    documento: str
    seccion: str
    texto: str
    vigente: bool

    @property
    def cita(self):
        return f"{self.documento}#{self.seccion}"


# 1. Trocear: un fragmento por cada sección "## " de cada documento,
#    con metadatos (documento, sección y si está vigente).
def trocear(carpeta=CARPETA):
    fragmentos = []
    for ruta in sorted(carpeta.glob("*.md")):
        texto = ruta.read_text(encoding="utf-8")
        vigente = "Estado: vigente" in texto
        for bloque in texto.split("\n## ")[1:]:
            titulo, _, cuerpo = bloque.partition("\n")
            fragmentos.append(Fragmento(ruta.stem, titulo.strip(), cuerpo.strip(), vigente))
    return fragmentos


# 2. Vectorizar: raíces de palabras sin tildes, pesadas por lo raras que son (TF-IDF).
def raices(texto):
    sin_tildes = unicodedata.normalize("NFD", texto.lower())
    sin_tildes = "".join(c for c in sin_tildes if unicodedata.category(c) != "Mn")
    return Counter(p[:5] for p in re.findall(r"[a-z0-9]+", sin_tildes) if p not in VACIAS)


def coseno(a, b):
    comun = sum(a[t] * b[t] for t in a if t in b)
    if not comun:
        return 0.0
    return comun / (math.sqrt(sum(v * v for v in a.values())) * math.sqrt(sum(v * v for v in b.values())))


class Indice:
    """Se construye una vez (la "indexación") y se consulta muchas."""

    def __init__(self, fragmentos):
        self.fragmentos = fragmentos
        recuentos = [raices(f.seccion + " " + f.texto) for f in fragmentos]
        apariciones = Counter(t for r in recuentos for t in r)
        n = len(fragmentos)
        self.idf = {t: math.log(1 + n / veces) for t, veces in apariciones.items()}
        self.vectores = [self._pesar(r) for r in recuentos]

    def _pesar(self, recuento):
        return {t: veces * self.idf.get(t, 0.0) for t, veces in recuento.items()}

    # 3. Recuperar: los k más parecidos que superen el umbral, con filtro por metadatos.
    def recuperar(self, pregunta, k=3, solo_vigentes=False):
        consulta = self._pesar(raices(pregunta))
        puntuados = [
            (coseno(consulta, vector), f)
            for vector, f in zip(self.vectores, self.fragmentos)
            if f.vigente or not solo_vigentes
        ]
        puntuados.sort(key=lambda par: par[0], reverse=True)
        return [(round(p, 2), f) for p, f in puntuados[:k] if p >= UMBRAL]


# 4. Aumentar: la pregunta viaja junto con los fragmentos y sus citas.
def construir_prompt(pregunta, recuperados):
    contexto = "\n\n".join(f"[{f.cita}]\n{f.texto}" for _, f in recuperados)
    return [
        {"role": "system", "content": (
            "Respondes dudas de los socios del club usando SOLO los fragmentos que se te dan. "
            "Cita entre corchetes el fragmento del que sale cada afirmación. "
            "Si los fragmentos no contienen la respuesta, di que no lo sabes.")},
        {"role": "user", "content": f"Fragmentos:\n{contexto}\n\nPregunta: {pregunta}"},
    ]


# 5. Generar, a través de un adaptador para no atar el código a un proveedor.
class ModeloSimulado:
    """Sin clave: devuelve la primera frase del mejor fragmento, con su cita."""

    def responder(self, mensajes):
        bloque = mensajes[-1]["content"].split("Fragmentos:\n", 1)[1].split("\n\n")[0]
        if not bloque.startswith("["):
            return "No lo sé: no he encontrado nada sobre eso en la documentación."
        cita, _, texto = bloque.partition("\n")
        return f"{texto.split('. ')[0].rstrip('.')}. {cita}"


class ModeloCompatibleOpenAI:
    def responder(self, mensajes):
        return completar(mensajes, simulador=None, temperatura=0.0)


def crear_modelo():
    return ModeloCompatibleOpenAI() if os.environ.get("LLM_API_KEY") else ModeloSimulado()


def preguntar(indice, pregunta, solo_vigentes=True, modelo=None):
    recuperados = indice.recuperar(pregunta, solo_vigentes=solo_vigentes)
    respuesta = (modelo or crear_modelo()).responder(construir_prompt(pregunta, recuperados))
    return respuesta, recuperados


if __name__ == "__main__":
    pregunta = sys.argv[1] if len(sys.argv) > 1 else "¿Cuántos invitados puedo llevar?"
    indice = Indice(trocear())
    for solo_vigentes in (False, True):
        respuesta, recuperados = preguntar(indice, pregunta, solo_vigentes)
        print(f"Pregunta: {pregunta}  (solo vigentes: {solo_vigentes})")
        for puntuacion, f in recuperados:
            print(f"  {puntuacion:.2f}  {f.cita}")
        print(f"Respuesta: {respuesta}\n")
