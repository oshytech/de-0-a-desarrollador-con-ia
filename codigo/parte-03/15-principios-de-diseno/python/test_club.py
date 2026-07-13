import unittest
from datetime import datetime

from club import RepositorioEnMemoria, ReservaRechazada, Socio, Solicitud, reservar

AHORA = datetime(2026, 3, 10, 12, 0)
LUCIA = Socio(1, "Lucía", "lucia@example.com", activo=True)


class TestReservar(unittest.TestCase):
    def setUp(self):
        self.repo = RepositorioEnMemoria()
        self.avisos = []
        self.notificar = lambda socio, mensaje: self.avisos.append((socio.id, mensaje))

    def pedir(self, socio, pista, inicio):
        return reservar(Solicitud(socio, pista, inicio), self.repo, self.notificar, AHORA)

    def test_reserva_valida_se_guarda_y_se_avisa(self):
        self.pedir(LUCIA, "Pádel 1", datetime(2026, 3, 11, 19, 0))
        self.assertEqual(len(self.repo.todas()), 1)
        self.assertEqual(len(self.avisos), 1)

    def test_no_se_reserva_una_pista_ocupada(self):
        marcos = Socio(2, "Marcos", "marcos@example.com", activo=True)
        self.pedir(LUCIA, "Pádel 1", datetime(2026, 3, 11, 19, 0))
        with self.assertRaises(ReservaRechazada):
            self.pedir(marcos, "Pádel 1", datetime(2026, 3, 11, 19, 0))

    def test_tercera_reserva_futura_rechazada(self):
        self.pedir(LUCIA, "Pádel 1", datetime(2026, 3, 11, 9, 0))
        self.pedir(LUCIA, "Pádel 1", datetime(2026, 3, 11, 10, 0))
        with self.assertRaises(ReservaRechazada):
            self.pedir(LUCIA, "Pádel 1", datetime(2026, 3, 11, 11, 0))
        self.assertEqual(len(self.avisos), 2)

    def test_socio_inactivo_no_reserva(self):
        baja = Socio(3, "Inés", "ines@example.com", activo=False)
        with self.assertRaises(ReservaRechazada):
            self.pedir(baja, "Tenis 1", datetime(2026, 3, 11, 9, 0))

    def test_ni_pasado_ni_mas_de_14_dias(self):
        for inicio in (datetime(2026, 3, 10, 11, 0), datetime(2026, 3, 25, 9, 0)):
            with self.assertRaises(ReservaRechazada):
                self.pedir(LUCIA, "Tenis 1", inicio)


if __name__ == "__main__":
    unittest.main()
