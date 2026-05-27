"""Primera versión del validador de DNI. Tiene un fallo a propósito:
no tiene en cuenta que la letra puede venir en minúscula."""

LETRAS = "TRWAGMYFPDXBNJZSQVHLCKE"


def dni_valido(texto):
    texto = texto.strip().replace("-", "").replace(" ", "")
    if len(texto) != 9:
        return False
    numero, letra = texto[:8], texto[8]
    if not numero.isdigit():
        return False
    return LETRAS[int(numero) % 23] == letra


if __name__ == "__main__":
    for prueba in ["12345678Z", "12345678-Z", "12345678z", "12345678A", "1234567Z"]:
        print(f"{prueba!r:14} -> {dni_valido(prueba)}")
