import unittest

from avisos import Notificador


class TransporteFalso:
    """Doble de prueba: no sale a la red, apunta cada petición que recibe."""

    def __init__(self):
        self.peticiones = []

    def post(self, ruta, cuerpo, cabeceras=None):
        self.peticiones.append(ruta)
        if ruta == "/token":
            return {"token": "token-de-prueba"}
        return {"ok": True}


class TestNotificador(unittest.TestCase):
    def test_envia_un_mensaje_por_recordatorio(self):
        transporte = TransporteFalso()
        notificador = Notificador(transporte, "club", "clave-de-mentira")

        notificador.recordar("ana@example.com", "pista 1", "10:00")

        self.assertEqual(transporte.peticiones.count("/mensajes"), 1)

    def test_pide_el_token_una_sola_vez(self):
        # Blinda una regresión que ningún test funcional ve: si alguien vuelve a
        # crear el cliente en cada envío, los avisos siguen llegando, pero el
        # tráfico contra la API se multiplica.
        transporte = TransporteFalso()
        notificador = Notificador(transporte, "club", "clave-de-mentira")

        for hora in ("10:00", "11:00", "12:00"):
            notificador.recordar("ana@example.com", "pista 1", hora)

        self.assertEqual(transporte.peticiones, ["/token", "/mensajes", "/mensajes", "/mensajes"])


if __name__ == "__main__":
    unittest.main()
