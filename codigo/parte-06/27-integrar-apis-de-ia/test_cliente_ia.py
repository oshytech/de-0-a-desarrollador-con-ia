"""Prueba el camino real del cliente contra un servidor falso que corre en local.

Así se comprueban la petición HTTP, los reintentos y la lectura del consumo
sin clave y sin gastar nada.

Uso: python -I test_cliente_ia.py
"""
import json
import os
import sys
import threading
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from unittest import mock

sys.path.insert(0, str(Path(__file__).parent))
from cliente_ia import ErrorIA, preguntar

RESPUESTA = {
    "id": "prueba-1",
    "choices": [{"message": {"role": "assistant", "content": "hola"}, "finish_reason": "stop"}],
    "usage": {"prompt_tokens": 12, "completion_tokens": 3},
}


class ServidorFalso(BaseHTTPRequestHandler):
    errores_pendientes = 0
    recibido = []

    def do_POST(self):
        cuerpo = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
        ServidorFalso.recibido.append((self.path, self.headers.get("Authorization"), cuerpo))
        if ServidorFalso.errores_pendientes:
            ServidorFalso.errores_pendientes -= 1
            self.send_response(503)
            self.end_headers()
            self.wfile.write(b'{"error": "saturado"}')
            return
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps(RESPUESTA).encode())

    def log_message(self, *args):
        pass


class CaminoReal(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.servidor = HTTPServer(("127.0.0.1", 0), ServidorFalso)
        threading.Thread(target=cls.servidor.serve_forever, daemon=True).start()
        cls.entorno = {"LLM_BASE_URL": f"http://127.0.0.1:{cls.servidor.server_port}/v1",
                       "LLM_API_KEY": "clave-de-mentira", "LLM_MODEL": "modelo-de-prueba"}

    @classmethod
    def tearDownClass(cls):
        cls.servidor.shutdown()

    def setUp(self):
        ServidorFalso.recibido.clear()
        ServidorFalso.errores_pendientes = 0

    def test_envia_el_formato_esperado_y_lee_el_consumo(self):
        with mock.patch.dict(os.environ, self.entorno):
            r = preguntar("sistema", "usuario", simulador=None, max_tokens=50)
        ruta, autorizacion, cuerpo = ServidorFalso.recibido[0]
        self.assertEqual(ruta, "/v1/chat/completions")
        self.assertEqual(autorizacion, "Bearer clave-de-mentira")
        self.assertEqual(cuerpo["model"], "modelo-de-prueba")
        self.assertEqual(cuerpo["max_tokens"], 50)
        self.assertEqual((r.texto, r.motivo_fin, r.tokens_entrada, r.tokens_salida, r.simulada),
                         ("hola", "stop", 12, 3, False))

    def test_reintenta_un_error_temporal(self):
        ServidorFalso.errores_pendientes = 1
        with mock.patch.dict(os.environ, self.entorno):
            r = preguntar("s", "u", simulador=None, espera=0.01)
        self.assertEqual(len(ServidorFalso.recibido), 2)
        self.assertEqual(r.texto, "hola")

    def test_se_rinde_cuando_se_acaban_los_reintentos(self):
        ServidorFalso.errores_pendientes = 5
        with mock.patch.dict(os.environ, self.entorno), self.assertRaises(ErrorIA):
            preguntar("s", "u", simulador=None, reintentos=2, espera=0.01)
        self.assertEqual(len(ServidorFalso.recibido), 3)

    def test_sin_modelo_configurado_no_llama(self):
        entorno = dict(self.entorno, LLM_MODEL="")
        with mock.patch.dict(os.environ, entorno), self.assertRaises(ErrorIA):
            preguntar("s", "u", simulador=None)
        self.assertEqual(ServidorFalso.recibido, [])


if __name__ == "__main__":
    unittest.main()
