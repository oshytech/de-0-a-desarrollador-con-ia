"""Strategy, Factory y Observer escritos con funciones, sin jerarquías de clases."""
import sys
from collections import defaultdict


# Strategy: cada forma de avisar es una función con la misma firma.
def por_correo(socio, mensaje):
    print(f"[correo a {socio}] {mensaje}")


def por_sms(socio, mensaje):
    print(f"[SMS a {socio}] {mensaje}")


# Factory: un diccionario elige la estrategia según la preferencia del socio.
CANALES = {"correo": por_correo, "sms": por_sms}


def canal_para(preferencia):
    return CANALES.get(preferencia, por_correo)


# Observer: quien publica un suceso no sabe quién lo escucha.
class Sucesos:
    def __init__(self):
        self._oyentes = defaultdict(list)

    def suscribir(self, suceso, oyente):
        self._oyentes[suceso].append(oyente)

    def publicar(self, suceso, **datos):
        for oyente in self._oyentes[suceso]:
            oyente(**datos)


if __name__ == "__main__":
    sys.stdout.reconfigure(encoding="utf-8")
    canal_para("sms")("Lucía", "Tu reserva es mañana a las 19:00")

    sucesos = Sucesos()
    sucesos.suscribir("reserva_cancelada",
                      lambda pista, hora: print(f"Avisar a la lista de espera de {pista} ({hora})"))
    sucesos.suscribir("reserva_cancelada",
                      lambda pista, hora: print(f"Registrar en el historial: {pista} libre a las {hora}"))
    sucesos.publicar("reserva_cancelada", pista="Pádel 1", hora="19:00")
