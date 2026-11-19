"""Uso: python -m asistente "¿Hasta cuándo puedo cancelar una reserva?" """

import sys
from pathlib import Path

from .asistente import Asistente, crear_componentes

CARPETA = Path(__file__).resolve().parent.parent / "reglamento"


def main(argumentos: list[str]) -> int:
    sys.stdout.reconfigure(encoding="utf-8")
    _, _, modo = crear_componentes()
    asistente = Asistente.desde_carpeta(CARPETA)
    preguntas = [" ".join(argumentos)] if argumentos else None
    if preguntas is None:
        print(f"Asistente del reglamento (modo {modo}). Línea vacía para salir.")
    while True:
        pregunta = preguntas.pop() if preguntas else (input("> ") if preguntas is None else "")
        if not pregunta.strip():
            return 0
        try:
            respuesta = asistente.preguntar(pregunta)
        except ValueError as error:
            print(error)
            continue
        print(respuesta.texto)
        if respuesta.fuentes:
            print("Fragmentos consultados:")
        for n, fuente in enumerate(respuesta.fuentes, start=1):
            print(f"  [{n}] {fuente.fuente}, sección \"{fuente.seccion}\"")


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
