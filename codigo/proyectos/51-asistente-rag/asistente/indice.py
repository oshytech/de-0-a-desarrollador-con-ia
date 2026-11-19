"""Índice vectorial en memoria con búsqueda por similitud coseno."""

import json
from dataclasses import asdict
from pathlib import Path

from .embeddings import ModeloEmbeddings
from .trocear import Fragmento


class Indice:
    def __init__(self, modelo: ModeloEmbeddings):
        self.modelo = modelo
        self.fragmentos: list[Fragmento] = []
        self.vectores: list[list[float]] = []

    def anadir(self, fragmentos: list[Fragmento]) -> None:
        # Se vectoriza sección + texto: el título ayuda a encontrar el fragmento.
        textos = [f"{f.seccion}. {f.texto}" for f in fragmentos]
        self.vectores += self.modelo.vectorizar(textos)
        self.fragmentos += fragmentos

    def buscar(self, pregunta: str, k: int = 3, umbral: float = 0.2) -> list[tuple[Fragmento, float]]:
        """Los `k` fragmentos más parecidos a la pregunta que superan el umbral."""
        [q] = self.modelo.vectorizar([pregunta])
        # Los vectores son unitarios, así que el producto escalar es la similitud coseno.
        puntuados = [(f, sum(a * b for a, b in zip(q, v))) for f, v in zip(self.fragmentos, self.vectores)]
        puntuados.sort(key=lambda par: par[1], reverse=True)
        return [(f, p) for f, p in puntuados[:k] if p >= umbral]

    def guardar(self, ruta: Path) -> None:
        datos = {
            "modelo": self.modelo.nombre,
            "fragmentos": [asdict(f) for f in self.fragmentos],
            "vectores": self.vectores,
        }
        ruta.write_text(json.dumps(datos, ensure_ascii=False), encoding="utf-8")

    @classmethod
    def cargar(cls, ruta: Path, modelo: ModeloEmbeddings) -> "Indice":
        datos = json.loads(ruta.read_text(encoding="utf-8"))
        if datos["modelo"] != modelo.nombre:
            # Pregunta y documentos tienen que pasar por el mismo modelo o los vectores no son comparables.
            raise ValueError(
                f"El índice se creó con '{datos['modelo']}' y ahora se usa '{modelo.nombre}'. Vuelve a indexar."
            )
        indice = cls(modelo)
        indice.fragmentos = [Fragmento(**f) for f in datos["fragmentos"]]
        indice.vectores = datos["vectores"]
        return indice
