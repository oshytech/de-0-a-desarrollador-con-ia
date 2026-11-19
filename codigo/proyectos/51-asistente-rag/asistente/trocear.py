"""Troceado de documentos Markdown en fragmentos con su sección de origen."""

from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class Fragmento:
    id: str
    fuente: str  # nombre del fichero
    seccion: str  # título de la sección, para citar
    texto: str


def _secciones(markdown: str) -> list[tuple[str, list[str]]]:
    """Agrupa los párrafos bajo el título "## ..." más reciente."""
    secciones: list[tuple[str, list[str]]] = []
    titulo = "Introducción"
    parrafos: list[str] = []
    for bloque in markdown.split("\n\n"):
        bloque = bloque.strip()
        if not bloque or bloque.startswith("# "):
            continue
        if bloque.startswith("## "):
            if parrafos:
                secciones.append((titulo, parrafos))
            titulo, parrafos = bloque[3:].strip(), []
        else:
            parrafos.append(" ".join(bloque.split()))
    if parrafos:
        secciones.append((titulo, parrafos))
    return secciones


def trocear(markdown: str, fuente: str, max_caracteres: int = 500) -> list[Fragmento]:
    """
    Corta por secciones y, dentro de cada una, junta párrafos enteros hasta
    `max_caracteres`. Nunca parte un párrafo por la mitad: en un reglamento,
    una regla cortada en dos es una regla que nadie va a encontrar entera.
    """
    fragmentos: list[Fragmento] = []
    for titulo, parrafos in _secciones(markdown):
        actual: list[str] = []
        for parrafo in parrafos:
            if actual and len(" ".join(actual + [parrafo])) > max_caracteres:
                fragmentos.append(_crear(fuente, titulo, actual, len(fragmentos)))
                actual = []
            actual.append(parrafo)
        if actual:
            fragmentos.append(_crear(fuente, titulo, actual, len(fragmentos)))
    return fragmentos


def _crear(fuente: str, titulo: str, parrafos: list[str], n: int) -> Fragmento:
    return Fragmento(id=f"{fuente}#{n}", fuente=fuente, seccion=titulo, texto=" ".join(parrafos))


def trocear_carpeta(carpeta: Path, max_caracteres: int = 500) -> list[Fragmento]:
    fragmentos: list[Fragmento] = []
    for ruta in sorted(carpeta.glob("*.md")):
        fragmentos += trocear(ruta.read_text(encoding="utf-8"), ruta.name, max_caracteres)
    return fragmentos
