"""Tests de las puertas, la taxonomía, el agregador y el presupuesto. Sin clave ni red.

Uso: python -I test_orquestacion.py
"""
import io
import json
import sys
import unittest
from contextlib import redirect_stdout
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import orquestacion as o

RESERVA = {"socio": "Lucía", "pista": "Pádel 2", "dia": "2026-12-03", "hora": "19:00"}


class Enrutar(unittest.TestCase):
    def test_una_respuesta_fuera_de_la_taxonomia_no_se_enruta(self):
        self.assertEqual(o.enrutar("lo que sea", simulador=lambda s, u: "Creo que es una reserva"), "FUERA")

    def test_acepta_la_palabra_aunque_venga_con_espacios(self):
        self.assertEqual(o.enrutar("x", simulador=lambda s, u: " reserva\n"), "RESERVA")


class CadenaConPuertas(unittest.TestCase):
    def test_puerta_1_corta_si_el_modelo_no_devuelve_json(self):
        with self.assertRaises(o.ErrorDePuerta) as e:
            o.cadena_reserva("pádel", extractor=lambda s, u: "Claro, te lo miro.")
        self.assertEqual(e.exception.puerta, 1)

    def test_puerta_1_corta_una_hora_fuera_del_horario(self):
        malo = lambda s, u: json.dumps({"deporte": "padel", "inicio": "2026-12-03T22:00"})
        with self.assertRaises(o.ErrorDePuerta) as e:
            o.cadena_reserva("pádel", extractor=malo)
        self.assertEqual(e.exception.puerta, 1)

    def test_puerta_3_corta_una_respuesta_que_cambia_la_hora(self):
        inventa = lambda s, u: "Te hemos reservado Pádel 2 el 2026-12-03 a las 20:00."
        with self.assertRaises(o.ErrorDePuerta) as e:
            o.cadena_reserva("Pádel el jueves a las 7 de la tarde", redactor=inventa)
        self.assertEqual(e.exception.puerta, 3)


class Paralelo(unittest.TestCase):
    def test_quita_duplicados_y_ordena_por_severidad(self):
        informe, fallidos = o.revisar_en_paralelo(o.CODIGO_A_REVISAR)
        self.assertEqual(fallidos, [])
        self.assertEqual(len(informe), 4)
        self.assertEqual([h["severidad"] for h in informe], ["alta", "alta", "media", "baja"])

    def test_un_revisor_que_no_devuelve_json_no_tumba_a_los_demas(self):
        def simulador(sistema, usuario):
            return "no sé" if "legibilidad" in sistema else "[]"
        informe, fallidos = o.revisar_en_paralelo("x = 1", simulador=simulador)
        self.assertEqual((informe, fallidos), ([], ["legibilidad"]))


class GenerarYEvaluar(unittest.TestCase):
    def test_pasa_cuando_el_generador_aplica_la_correccion(self):
        with redirect_stdout(io.StringIO()):
            _, vueltas = o.generar_y_evaluar(RESERVA)
        self.assertEqual(vueltas, 2)

    def test_escala_a_una_persona_al_agotar_el_presupuesto(self):
        llamadas = []

        def terco(sistema, usuario):
            llamadas.append(1)
            return "Tu reserva queda garantizada."
        with redirect_stdout(io.StringIO()), self.assertRaises(o.Escalar):
            o.generar_y_evaluar(RESERVA, generador=terco, max_vueltas=3)
        self.assertEqual(len(llamadas), 3)

    def test_el_evaluador_detecta_promesas(self):
        self.assertIn("Promete algo que el club no ha decidido.",
                      o.evaluar("Pádel 2, 2026-12-03, 19:00. Te lo garantizamos.", RESERVA))


if __name__ == "__main__":
    unittest.main()
