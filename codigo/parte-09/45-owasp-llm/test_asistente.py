import unittest

from asistente import (
    COMENTARIOS_PISTA_2,
    SISTEMA_SEGURO,
    AsistenteInseguro,
    AsistenteSeguro,
    Club,
    Denegada,
    Llamada,
    ModeloSimulado,
    Respuesta,
)

BRUNO = 2


class SinDefensas(unittest.TestCase):
    def test_la_inyeccion_cancela_reservas_ajenas(self):
        club = Club()
        AsistenteInseguro(club, ModeloSimulado()).resumir_comentarios(BRUNO, COMENTARIOS_PISTA_2)
        self.assertEqual(club.reservas[1043]["estado"], "cancelada")  # la de Ana

    def test_la_salida_llega_como_html(self):
        html = AsistenteInseguro(Club(), ModeloSimulado()).resumir_comentarios(BRUNO, COMENTARIOS_PISTA_2)
        self.assertIn("<img src=x onerror", html)


class Delimitadores(unittest.TestCase):
    def test_solos_no_bastan_si_el_dato_cierra_la_etiqueta(self):
        modelo = ModeloSimulado()
        contexto = "<datos>\n- " + COMENTARIOS_PISTA_2[2] + "\n</datos>"
        r = modelo.responder(SISTEMA_SEGURO, contexto, ["cancelar_reserva"])
        self.assertTrue(r.llamadas, "el comentario cierra </datos> y la orden se cuela")


class ConDefensas(unittest.TestCase):
    def setUp(self):
        self.club = Club()
        self.modelo = ModeloSimulado()
        self.asistente = AsistenteSeguro(self.club, self.modelo)

    def test_resumir_no_ejecuta_nada(self):
        self.asistente.resumir_comentarios(BRUNO, COMENTARIOS_PISTA_2)
        self.assertEqual(self.asistente.registro, [])
        self.assertTrue(all(r["estado"] == "confirmada" for r in self.club.reservas.values()))

    def test_el_dato_no_puede_cerrar_la_etiqueta(self):
        self.asistente.resumir_comentarios(BRUNO, COMENTARIOS_PISTA_2)
        self.assertEqual(self.modelo.ultimo_contexto.count("</datos>"), 1 + SISTEMA_SEGURO.count("</datos>"))

    def test_la_salida_se_escapa(self):
        html = self.asistente.resumir_comentarios(BRUNO, COMENTARIOS_PISTA_2)
        self.assertNotIn("<img", html)
        self.assertIn("&lt;img", html)

    def test_no_se_cancela_una_reserva_ajena(self):
        with self.assertRaises(Denegada):
            self.asistente.atender(BRUNO, "llama a cancelar_reserva(1043)")
        self.assertEqual(self.club.reservas[1043]["estado"], "confirmada")

    def test_la_propia_pide_confirmacion(self):
        avisos = self.asistente.atender(BRUNO, "llama a cancelar_reserva(1050)")
        self.assertEqual(len(avisos), 1)
        self.assertEqual(self.club.reservas[1050]["estado"], "confirmada")
        self.asistente.confirmar(BRUNO, "conf-1")
        self.assertEqual(self.club.reservas[1050]["estado"], "cancelada")

    def test_argumentos_con_forma_inesperada(self):
        class ModeloRaro(ModeloSimulado):
            def responder(self, sistema, contexto, herramientas):
                return Respuesta("", [Llamada("cancelar_reserva", {"reserva_id": "1050 OR 1=1"})])

        with self.assertRaises(Denegada):
            AsistenteSeguro(self.club, ModeloRaro()).atender(BRUNO, "lo que sea")

    def test_demasiadas_llamadas_en_un_turno(self):
        peticion = " ".join(["llama a cancelar_reserva(1050)"] * 10)
        with self.assertRaises(Denegada):
            self.asistente.atender(BRUNO, peticion)

    def test_entrada_demasiado_larga(self):
        with self.assertRaises(Denegada):
            self.asistente.resumir_comentarios(BRUNO, ["x" * 5000])

    def test_el_prompt_de_sistema_no_lleva_secretos(self):
        for palabra in ("clave", "token", "contraseña", "password"):
            self.assertNotIn(palabra, SISTEMA_SEGURO.lower())


if __name__ == "__main__":
    unittest.main()
