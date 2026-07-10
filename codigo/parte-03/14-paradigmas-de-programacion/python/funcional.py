"""Estilo funcional en Python: datos inmutables, funciones puras y funciones como valores."""
from dataclasses import dataclass, replace
from datetime import datetime, timedelta


@dataclass(frozen=True)
class Reserva:
    socio: str
    pista: str
    inicio: datetime
    estado: str = "confirmada"


def cancelada(reserva):
    """No modifica la reserva: devuelve una copia con otro estado."""
    return replace(reserva, estado="cancelada")


def futuras_confirmadas(reservas, socio, ahora):
    return [r for r in reservas
            if r.socio == socio and r.estado == "confirmada" and r.inicio > ahora]


def puede_cancelar_con_margen(horas):
    """Devuelve una función: cada club podría fijar su propio margen."""
    margen = timedelta(hours=horas)
    return lambda reserva, ahora: reserva.inicio - ahora >= margen


if __name__ == "__main__":
    ahora = datetime(2026, 3, 10, 12, 0)
    reservas = [
        Reserva("Lucía", "Tenis 1", datetime(2026, 3, 9, 18, 0)),
        Reserva("Lucía", "Pádel 2", datetime(2026, 3, 11, 9, 0)),
        Reserva("Marcos", "Pádel 2", datetime(2026, 3, 11, 10, 0)),
        cancelada(Reserva("Lucía", "Tenis 2", datetime(2026, 3, 12, 20, 0))),
    ]
    print(len(futuras_confirmadas(reservas, "Lucía", ahora)))

    puede_cancelar = puede_cancelar_con_margen(2)
    print(puede_cancelar(reservas[1], ahora))
    print(puede_cancelar(reservas[1], datetime(2026, 3, 11, 8, 0)))

    original = reservas[2]
    copia = cancelada(original)
    print(original.estado, copia.estado)
