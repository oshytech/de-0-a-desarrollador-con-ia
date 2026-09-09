"""Tests de la validación y de las reglas. No necesitan clave ni red.

Uso: python -I test_peticion_reserva.py
"""
import sys
import unittest
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from peticion_reserva import ErrorValidacion, decidir, extraer_json, procesar, validar_peticion

AHORA = datetime(2026, 12, 1, 10, 0)


def peticion(**cambios):
    base = {"deporte": "padel", "fecha": "2026-12-02", "hora": "10:00", "dudas": []}
    base.update(cambios)
    return base


class ValidarSalidaDelModelo(unittest.TestCase):
    def test_acepta_json_entre_marcas_de_codigo(self):
        self.assertEqual(extraer_json('```json\n{"a": 1}\n```'), {"a": 1})

    def test_rechaza_texto_sin_json(self):
        with self.assertRaises(ErrorValidacion):
            extraer_json("Claro, aquí tienes tu reserva.")

    def test_rechaza_claves_de_mas_o_de_menos(self):
        with self.assertRaises(ErrorValidacion):
            validar_peticion({"deporte": "padel", "fecha": "2026-12-02", "hora": "10:00"})

    def test_rechaza_deporte_fuera_del_catalogo(self):
        with self.assertRaises(ErrorValidacion):
            validar_peticion(peticion(deporte="squash"))

    def test_rechaza_hora_en_formato_de_12_horas(self):
        with self.assertRaises(ErrorValidacion):
            validar_peticion(peticion(hora="7 pm"))


class ReglasDelClub(unittest.TestCase):
    def test_la_ultima_reserva_empieza_a_las_21(self):
        self.assertEqual(decidir(peticion(hora="21:00"), AHORA)[0], "consultar")
        self.assertEqual(decidir(peticion(hora="22:00"), AHORA)[0], "rechazar")

    def test_solo_en_punto(self):
        self.assertEqual(decidir(peticion(hora="10:30"), AHORA)[0], "rechazar")

    def test_ni_en_el_pasado_ni_a_mas_de_14_dias(self):
        self.assertEqual(decidir(peticion(fecha="2026-12-01", hora="09:00"), AHORA)[0], "rechazar")
        self.assertEqual(decidir(peticion(fecha="2026-12-15", hora="10:00"), AHORA)[0], "consultar")
        self.assertEqual(decidir(peticion(fecha="2026-12-15", hora="11:00"), AHORA)[0], "rechazar")

    def test_si_falta_un_dato_se_pide_aclaracion(self):
        accion, motivo = decidir(peticion(hora=None, dudas=["No se indica la hora."]), AHORA)
        self.assertEqual((accion, motivo), ("aclarar", "No se indica la hora."))


class DeExtremoAExtremoEnModoSimulado(unittest.TestCase):
    def test_un_mensaje_completo(self):
        datos, (accion, _), respuesta = procesar("¿Me reservas pádel el jueves a las 7 de la tarde?")
        self.assertTrue(respuesta.simulada)
        self.assertEqual((datos["fecha"], datos["hora"], accion), ("2026-12-03", "19:00", "consultar"))


if __name__ == "__main__":
    unittest.main()
