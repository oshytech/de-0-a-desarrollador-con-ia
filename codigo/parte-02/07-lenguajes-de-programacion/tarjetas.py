"""Orientación a objetos: datos y comportamiento juntos."""

PRECIO_VIAJE = 150  # en céntimos


class TarjetaTransporte:
    def __init__(self, titular):
        self.titular = titular
        self.saldo = 0  # en céntimos

    def recargar(self, cantidad):
        if cantidad <= 0:
            raise ValueError("La recarga tiene que ser positiva")
        self.saldo += cantidad

    def precio_viaje(self):
        return PRECIO_VIAJE

    def validar_viaje(self):
        precio = self.precio_viaje()
        if self.saldo < precio:
            raise ValueError(f"Saldo insuficiente para {self.titular}")
        self.saldo -= precio


class TarjetaJoven(TarjetaTransporte):
    def precio_viaje(self):
        return PRECIO_VIAJE // 2


tarjetas = [TarjetaTransporte("Marta"), TarjetaJoven("Iker")]
for tarjeta in tarjetas:
    tarjeta.recargar(1000)
    tarjeta.validar_viaje()
    print(f"{tarjeta.titular}: quedan {tarjeta.saldo / 100:.2f} euros")
