import itertools
import unittest

import informe
import informe_antes

PISTAS = [{"id": "p1", "nombre": "Pista 1"}, {"id": "p2", "nombre": "Pista 2"}]


def reservas_de_prueba(confirmadas_p1, canceladas_p1, confirmadas_p2):
    return (
        [{"pista": "p1", "estado": "confirmada"}] * confirmadas_p1
        + [{"pista": "p1", "estado": "cancelada"}] * canceladas_p1
        + [{"pista": "p2", "estado": "confirmada"}] * confirmadas_p2
    )


class TestParidadInforme(unittest.TestCase):
    def test_misma_salida_que_la_version_antigua(self):
        for c1, x1, c2 in itertools.product(range(15), range(3), range(15)):
            reservas = reservas_de_prueba(c1, x1, c2)
            with self.subTest(c1=c1, x1=x1, c2=c2):
                self.assertEqual(informe.informe(reservas, PISTAS), informe_antes.informe(reservas, PISTAS))


class TestInforme(unittest.TestCase):
    def test_las_canceladas_no_cuentan(self):
        self.assertEqual(informe.horas_confirmadas(reservas_de_prueba(3, 2, 0), "p1"), 3)

    def test_marca_la_pista_llena_por_encima_del_umbral(self):
        self.assertEqual(informe.linea_de_pista("Pista 1", 78.6), "Pista 1: 79% (llena)")
        self.assertEqual(informe.linea_de_pista("Pista 1", 75), "Pista 1: 75%")


if __name__ == "__main__":
    unittest.main()
