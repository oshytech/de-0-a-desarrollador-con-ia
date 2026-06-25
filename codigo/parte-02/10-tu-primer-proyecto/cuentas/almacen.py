"""Guardar y cargar los gastos en un fichero JSON."""

import json


class FicheroDanado(Exception):
    """El fichero existe pero no contiene un JSON válido."""


def cargar(ruta):
    """Devuelve la lista de gastos guardada, o una lista vacía si no hay fichero."""
    try:
        with open(ruta, encoding="utf-8") as fichero:
            return json.load(fichero)
    except FileNotFoundError:
        return []
    except json.JSONDecodeError as error:
        raise FicheroDanado(
            f"{ruta} no es un JSON válido (línea {error.lineno})"
        ) from error


def guardar(ruta, gastos):
    """Escribe la lista completa de gastos."""
    with open(ruta, "w", encoding="utf-8") as fichero:
        json.dump(gastos, fichero, ensure_ascii=False, indent=2)
