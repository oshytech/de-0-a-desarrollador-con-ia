"""Une las piezas: configuración, índice, búsqueda y generación."""

import os
from dataclasses import dataclass
from pathlib import Path

from .embeddings import EmbeddingsLocales, EmbeddingsRemotos, ModeloEmbeddings
from .generador import SIN_RESPUESTA, Generador, GeneradorRemoto, GeneradorSimulado
from .indice import Indice
from .trocear import Fragmento, trocear_carpeta

MAX_PREGUNTA = 500


@dataclass
class Respuesta:
    texto: str
    fuentes: list[Fragmento]


def crear_componentes(entorno: dict[str, str] | None = None) -> tuple[ModeloEmbeddings, Generador, str]:
    """
    Sin ASISTENTE_API_KEY (o con ASISTENTE_MODO=simulado) todo funciona en local, sin red.
    Con clave, usa cualquier proveedor compatible con la API de OpenAI, incluido un servidor
    local de modelos que exponga esa misma API.
    """
    entorno = dict(os.environ) if entorno is None else entorno
    clave = entorno.get("ASISTENTE_API_KEY", "")
    if not clave or entorno.get("ASISTENTE_MODO") == "simulado":
        return EmbeddingsLocales(), GeneradorSimulado(), "simulado"
    url = entorno.get("ASISTENTE_API_URL", "https://api.openai.com/v1")
    return (
        EmbeddingsRemotos(url, clave, entorno.get("ASISTENTE_MODELO_EMBEDDINGS", "text-embedding-3-small")),
        GeneradorRemoto(url, clave, entorno.get("ASISTENTE_MODELO", "gpt-4o-mini")),
        "remoto",
    )


class Asistente:
    def __init__(self, indice: Indice, generador: Generador, k: int = 3, umbral: float = 0.2):
        self.indice, self.generador, self.k, self.umbral = indice, generador, k, umbral

    @classmethod
    def desde_carpeta(cls, carpeta: Path, entorno: dict[str, str] | None = None) -> "Asistente":
        modelo, generador, _ = crear_componentes(entorno)
        indice = Indice(modelo)
        indice.anadir(trocear_carpeta(carpeta))
        return cls(indice, generador)

    def preguntar(self, pregunta: str) -> Respuesta:
        pregunta = " ".join(pregunta.split())
        if not pregunta:
            raise ValueError("La pregunta está vacía.")
        if len(pregunta) > MAX_PREGUNTA:
            raise ValueError(f"La pregunta es demasiado larga (máximo {MAX_PREGUNTA} caracteres).")
        encontrados = [f for f, _ in self.indice.buscar(pregunta, self.k, self.umbral)]
        if not encontrados:
            # Sin contexto no se llama al modelo: así no tiene ocasión de inventar.
            return Respuesta(SIN_RESPUESTA, [])
        return Respuesta(self.generador.responder(pregunta, encontrados), encontrados)
