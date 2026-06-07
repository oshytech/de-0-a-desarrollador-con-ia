"""Convertir lo que escribe una persona en un número de pasos válido."""


def leer_pasos(texto):
    """Convierte un texto en un número de pasos. Lanza ValueError si no es válido."""
    pasos = int(texto)
    if pasos < 0:
        raise ValueError(f"Los pasos no pueden ser negativos: {pasos}")
    return pasos


for entrada in ["9500", " 12000 ", "nueve mil", "-20"]:
    try:
        print(f"{entrada!r} -> {leer_pasos(entrada)}")
    except ValueError as error:
        print(f"{entrada!r} -> no válido ({error})")
