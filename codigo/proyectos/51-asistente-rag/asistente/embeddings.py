"""Modelos de embeddings: uno local y simple, y otro remoto compatible con la API de OpenAI."""

import hashlib
import json
import math
import re
import unicodedata
import urllib.request
from typing import Protocol

PALABRAS_VACIAS = set(
    "a al algo ante como con de del desde donde el en entre era es esta este esto hay "
    "la las le lo los mas me mi mis muy no o para pero por que se si sin so su sus "
    "un una uno unos unas y ya yo tu te ti cual cuales cuando cuanto cuanta cuantos cuantas "
    "puedo puede pueden hacer tengo tiene hasta ser".split()
)


class ModeloEmbeddings(Protocol):
    nombre: str

    def vectorizar(self, textos: list[str]) -> list[list[float]]: ...


def normalizar(texto: str) -> list[str]:
    """Minúsculas, sin tildes, sin palabras vacías y con las palabras recortadas a su raíz."""
    sin_tildes = unicodedata.normalize("NFKD", texto.lower()).encode("ascii", "ignore").decode()
    palabras = re.findall(r"[a-z0-9]+", sin_tildes)
    # Raíz "a lo bruto": los 5 primeros caracteres. cancelar, cancelación y cancelaciones dan "cance".
    return [p[:5] for p in palabras if p not in PALABRAS_VACIAS]


class EmbeddingsLocales:
    """
    Bolsa de raíces con el truco del hashing: cada raíz suma 1 en una de `dimensiones`
    posiciones. No entiende sinónimos, pero no necesita red ni clave y siempre da lo mismo.
    Se usa hashlib y no hash(), porque hash() cambia entre ejecuciones de Python.
    """

    nombre = "local-hashing-v1"

    def __init__(self, dimensiones: int = 1024):
        self.dimensiones = dimensiones

    def _posicion(self, raiz: str) -> int:
        return int(hashlib.md5(raiz.encode()).hexdigest(), 16) % self.dimensiones

    def vectorizar(self, textos: list[str]) -> list[list[float]]:
        vectores = []
        for texto in textos:
            v = [0.0] * self.dimensiones
            for raiz in normalizar(texto):
                v[self._posicion(raiz)] += 1.0
            vectores.append(_unitario(v))
        return vectores


class EmbeddingsRemotos:
    """Llama a POST {url}/embeddings de cualquier proveedor compatible con la API de OpenAI."""

    def __init__(self, url: str, clave: str, modelo: str):
        self.url, self.clave, self.nombre = url.rstrip("/"), clave, modelo

    def vectorizar(self, textos: list[str]) -> list[list[float]]:
        datos = _post_json(f"{self.url}/embeddings", self.clave, {"model": self.nombre, "input": textos})
        return [_unitario(d["embedding"]) for d in sorted(datos["data"], key=lambda d: d["index"])]


def _unitario(v: list[float]) -> list[float]:
    norma = math.sqrt(sum(x * x for x in v))
    return [x / norma for x in v] if norma else v


def _post_json(url: str, clave: str, cuerpo: dict) -> dict:
    peticion = urllib.request.Request(
        url,
        data=json.dumps(cuerpo).encode(),
        headers={"Authorization": f"Bearer {clave}", "Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(peticion, timeout=60) as respuesta:
        return json.load(respuesta)
