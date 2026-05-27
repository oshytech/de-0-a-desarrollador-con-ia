"""Validador de DNI y NIE españoles, traducido del pseudocódigo del capítulo 4."""

LETRAS = "TRWAGMYFPDXBNJZSQVHLCKE"
PREFIJOS_NIE = {"X": "0", "Y": "1", "Z": "2"}


def limpiar(texto):
    """Quita espacios, guiones y puntos y pasa a mayúsculas."""
    for separador in " -.":
        texto = texto.replace(separador, "")
    return texto.upper()


def documento_valido(texto):
    """Devuelve True si el texto es un DNI o un NIE con la letra correcta."""
    texto = limpiar(texto)
    if len(texto) != 9:
        return False
    if texto[0] in PREFIJOS_NIE:
        texto = PREFIJOS_NIE[texto[0]] + texto[1:]
    numero, letra = texto[:8], texto[8]
    if not numero.isdigit():
        return False
    return LETRAS[int(numero) % 23] == letra


if __name__ == "__main__":
    casos = {
        "12345678Z": True,
        "12345678-Z": True,
        "12345678z": True,
        " 12345678 Z ": True,
        "12.345.678-Z": True,
        "12345678A": False,
        "1234567Z": False,
        "1234A678Z": False,
        "X1234567L": True,
        "X1234567T": False,
        "": False,
    }
    for texto, esperado in casos.items():
        obtenido = documento_valido(texto)
        marca = "ok" if obtenido == esperado else "FALLO"
        print(f"{texto!r:16} -> {obtenido!s:5} ({marca})")
