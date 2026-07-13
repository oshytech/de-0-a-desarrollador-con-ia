"""KISS: qué se pinta en cada casilla del calendario de una pista."""
import sys
from datetime import datetime, timedelta

ANTELACION_MAXIMA = timedelta(days=14)


def estado_casilla_enredado(inicio, reserva, socio_id, ahora):
    if inicio > ahora:
        if inicio - ahora <= ANTELACION_MAXIMA:
            if reserva is None:
                return "libre"
            else:
                if reserva["socio"] == socio_id:
                    return "tuya"
                else:
                    return "ocupada"
        else:
            return "aún no disponible"
    else:
        return "pasada"


def estado_casilla(inicio, reserva, socio_id, ahora):
    if inicio <= ahora:
        return "pasada"
    if inicio - ahora > ANTELACION_MAXIMA:
        return "aún no disponible"
    if reserva is None:
        return "libre"
    if reserva["socio"] == socio_id:
        return "tuya"
    return "ocupada"


if __name__ == "__main__":
    sys.stdout.reconfigure(encoding="utf-8")
    ahora = datetime(2026, 3, 10, 12, 0)
    casos = [
        (datetime(2026, 3, 10, 9, 0), None),
        (datetime(2026, 3, 30, 9, 0), None),
        (datetime(2026, 3, 11, 9, 0), None),
        (datetime(2026, 3, 11, 10, 0), {"socio": 1}),
        (datetime(2026, 3, 11, 11, 0), {"socio": 2}),
    ]
    for inicio, reserva in casos:
        a = estado_casilla_enredado(inicio, reserva, 1, ahora)
        b = estado_casilla(inicio, reserva, 1, ahora)
        assert a == b
        print(f"{inicio:%d/%m %H:%M} {b}")
