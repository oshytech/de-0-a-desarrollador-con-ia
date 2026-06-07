"""Antipatrón: capturar cualquier error y no hacer nada con él."""


def procesar_semana():
    pasos = {"lunes": 8250}
    return pasos["martes"]  # KeyError: el martes no existe


try:
    procesar_semana()
except Exception:
    pass  # no hagas esto

print("El programa sigue como si nada, y el KeyError ha desaparecido")
