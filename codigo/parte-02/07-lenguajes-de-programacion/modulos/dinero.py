"""Funciones para trabajar con importes guardados en céntimos."""


def a_centimos(texto):
    """Convierte "12,50" o "12.50" en 1250."""
    euros, _, centimos = texto.replace(",", ".").partition(".")
    return int(euros) * 100 + int(centimos.ljust(2, "0")[:2])


def formatear(centimos):
    """Convierte 1250 en "12,50 €"."""
    return f"{centimos // 100},{centimos % 100:02d} €"


if __name__ == "__main__":
    # Solo se ejecuta si lanzas este fichero directamente, no al importarlo
    print(formatear(a_centimos("12,5")))
