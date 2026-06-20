import unittest

from entrenamientos import (
    informe_semanal,
    mejor_sesion,
    ritmo,
    sesiones,
    sesiones_completadas,
)


class TestEntrenamientos(unittest.TestCase):
    def test_ritmo_en_minutos_por_km(self):
        self.assertEqual(ritmo(50, 10), 5)

    def test_descarta_las_sesiones_no_completadas(self):
        self.assertEqual(len(sesiones_completadas(sesiones)), 3)

    def test_mejor_sesion_es_la_de_ritmo_mas_rapido(self):
        self.assertEqual(
            mejor_sesion(sesiones_completadas(sesiones))["fecha"], "2026-03-09"
        )

    def test_el_informe_no_cambia_tras_el_refactor(self):
        esperado = (
            "Total: 23.7 km\n"
            "Ritmo medio: 5.65 min/km\n"
            "Mejor ritmo: 5.52 min/km el 2026-03-09\n"
            "Esta semana has corrido más que una media maratón"
        )
        self.assertEqual(informe_semanal(sesiones), esperado)


if __name__ == "__main__":
    unittest.main()
