import io
import os
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout

import almacen
import reparto
from main import main


class TestImportes(unittest.TestCase):
    def test_acepta_coma_punto_y_enteros(self):
        self.assertEqual(reparto.a_centimos("12,5"), 1250)
        self.assertEqual(reparto.a_centimos("12.50"), 1250)
        self.assertEqual(reparto.a_centimos(" 12 "), 1200)

    def test_rechaza_lo_que_no_es_un_importe(self):
        for texto in ["", "doce", "-5", "12,345", "0", "1.234,50"]:
            with self.subTest(texto=texto):
                with self.assertRaises(reparto.ImporteInvalido):
                    reparto.a_centimos(texto)

    def test_formatea_positivos_y_negativos(self):
        self.assertEqual(reparto.formatear(1250), "12,50 €")
        self.assertEqual(reparto.formatear(-320), "-3,20 €")
        self.assertEqual(reparto.formatear(5), "0,05 €")


class TestReparto(unittest.TestCase):
    def test_no_se_pierde_ningun_centimo(self):
        partes = reparto.repartir(1000, ["Ana", "Bruno", "Carla"])
        self.assertEqual(partes, {"Ana": 334, "Bruno": 333, "Carla": 333})
        self.assertEqual(sum(partes.values()), 1000)

    def test_una_persona_repetida_es_un_error(self):
        with self.assertRaises(ValueError):
            reparto.repartir(1000, ["Ana", "Ana"])

    def test_sin_personas_es_un_error(self):
        with self.assertRaises(ValueError):
            reparto.repartir(1000, [])

    def test_los_saldos_suman_cero(self):
        gastos = [
            {"pagado_por": "Ana", "importe": 6340, "entre": ["Ana", "Bruno", "Carla"]},
            {"pagado_por": "Carla", "importe": 4500, "entre": ["Ana", "Carla"]},
        ]
        saldos = reparto.calcular_saldos(gastos)
        self.assertEqual(saldos, {"Ana": 1976, "Bruno": -2113, "Carla": 137})
        self.assertEqual(sum(saldos.values()), 0)


class TestLiquidar(unittest.TestCase):
    def test_quien_mas_debe_paga_a_quien_mas_recibe(self):
        pagos = reparto.liquidar({"Ana": 20302, "Bruno": -8106, "Carla": -12196})
        self.assertEqual(pagos, [("Carla", "Ana", 12196), ("Bruno", "Ana", 8106)])

    def test_tras_los_pagos_todos_quedan_a_cero(self):
        saldos = {"Ana": 500, "Bruno": 300, "Carla": -650, "Dani": -150}
        for deudor, acreedor, cantidad in reparto.liquidar(saldos):
            saldos[deudor] += cantidad
            saldos[acreedor] -= cantidad
        self.assertTrue(all(saldo == 0 for saldo in saldos.values()))

    def test_saldos_que_no_cuadran_son_un_error(self):
        with self.assertRaises(ValueError):
            reparto.liquidar({"Ana": 100, "Bruno": -50})


class TestProgramaCompleto(unittest.TestCase):
    def setUp(self):
        carpeta = tempfile.mkdtemp()
        self.fichero = os.path.join(carpeta, "cuentas.json")

    def ejecutar(self, *argumentos):
        """Lanza el programa con un fichero temporal y sin ensuciar la salida."""
        with redirect_stdout(io.StringIO()), redirect_stderr(io.StringIO()):
            return main(["--fichero", self.fichero, *argumentos])

    def test_sin_fichero_no_hay_gastos(self):
        self.assertEqual(almacen.cargar(self.fichero), [])

    def test_anadir_guarda_el_gasto(self):
        codigo = self.ejecutar(
            "anadir", "Cena", "45", "--paga", "Ana", "--entre", "Ana", "Bruno"
        )
        self.assertEqual(codigo, 0)
        gastos = almacen.cargar(self.fichero)
        self.assertEqual(gastos[0]["importe"], 4500)

    def test_un_importe_invalido_no_se_guarda(self):
        codigo = self.ejecutar(
            "anadir", "Cena", "cuarenta", "--paga", "Ana", "--entre", "Ana"
        )
        self.assertEqual(codigo, 1)
        self.assertEqual(almacen.cargar(self.fichero), [])

    def test_un_fichero_danado_se_avisa(self):
        with open(self.fichero, "w", encoding="utf-8") as fichero:
            fichero.write("{esto no es json")
        with self.assertRaises(almacen.FicheroDanado):
            almacen.cargar(self.fichero)


if __name__ == "__main__":
    unittest.main()
