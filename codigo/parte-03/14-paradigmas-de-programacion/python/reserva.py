"""Orientación a objetos en Python: la reserva guarda su estado y protege sus reglas."""
from datetime import datetime, timedelta

MARGEN_CANCELACION = timedelta(hours=2)


class Reserva:
    def __init__(self, id, socio, pista, inicio):
        self.id = id
        self.socio = socio
        self.pista = pista
        self.inicio = inicio
        self.fin = inicio + timedelta(hours=1)
        self._estado = "confirmada"  # el guion bajo avisa: no lo toques desde fuera

    @property
    def estado(self):
        return self._estado

    def cancelar(self, ahora):
        if self._estado == "cancelada":
            raise ValueError("la reserva ya estaba cancelada")
        if self.inicio - ahora < MARGEN_CANCELACION:
            raise ValueError("solo se puede cancelar hasta 2 horas antes")
        self._estado = "cancelada"


if __name__ == "__main__":
    r = Reserva(1, "Lucía", "Pádel 1", datetime(2026, 3, 10, 19, 0))
    try:
        r.cancelar(ahora=datetime(2026, 3, 10, 17, 30))
    except ValueError as e:
        print("No:", e)
    r.cancelar(ahora=datetime(2026, 3, 10, 16, 0))
    print(r.estado)
