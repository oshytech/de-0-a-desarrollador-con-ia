"""Hace de cliente MCP contra servidor_mcp_club.py y comprueba sus respuestas.

Es lo mismo que hace un agente al conectarse: lanza el proceso, inicializa,
pide el catálogo de herramientas y las llama. Termina con código 1 si alguna
comprobación falla.

Uso: python -I probar_servidor.py
"""
import json
import os
import subprocess
import sys
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8")
SERVIDOR = Path(__file__).parent / "servidor_mcp_club.py"


class Sesion:
    def __init__(self, escritura=False):
        entorno = dict(os.environ, CLUB_MCP_ESCRITURA="1" if escritura else "0")
        self.proceso = subprocess.Popen(
            [sys.executable, "-I", str(SERVIDOR)],
            stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, env=entorno)
        self.siguiente_id = 1

    def enviar(self, mensaje):
        self.proceso.stdin.write(json.dumps(mensaje).encode("utf-8") + b"\n")
        self.proceso.stdin.flush()

    def pedir(self, metodo, params=None):
        mensaje = {"jsonrpc": "2.0", "id": self.siguiente_id, "method": metodo}
        if params is not None:
            mensaje["params"] = params
        self.siguiente_id += 1
        self.enviar(mensaje)
        return json.loads(self.proceso.stdout.readline())

    def iniciar(self):
        r = self.pedir("initialize", {"protocolVersion": "2025-06-18", "capabilities": {},
                                      "clientInfo": {"name": "probador", "version": "1.0"}})
        self.enviar({"jsonrpc": "2.0", "method": "notifications/initialized"})
        return r["result"]

    def herramienta(self, nombre, **argumentos):
        r = self.pedir("tools/call", {"name": nombre, "arguments": argumentos})
        resultado = r["result"]
        return resultado["isError"], resultado["content"][0]["text"]

    def cerrar(self):
        self.proceso.stdin.close()
        self.proceso.wait(timeout=5)


fallos = 0


def comprobar(descripcion, condicion):
    global fallos
    print(("OK    " if condicion else "FALLO ") + descripcion)
    fallos += 0 if condicion else 1


# Sesión 1: servidor por defecto, solo lectura
s = Sesion()
info = s.iniciar()
comprobar("initialize negocia la versión pedida", info["protocolVersion"] == "2025-06-18")
herramientas = [h["name"] for h in s.pedir("tools/list")["result"]["tools"]]
print("      herramientas:", ", ".join(herramientas))
comprobar("el catálogo anuncia cuatro herramientas", len(herramientas) == 4)

es_error, texto = s.herramienta("consultar_disponibilidad", pista_id="P2", fecha="2026-12-01")
libres = json.loads(texto)["horas_libres"]
comprobar("las 18:00 del 1 de diciembre no salen libres en Pádel 1", "18:00" not in libres)
comprobar("hay 13 horas libres ese día", len(libres) == 13)

es_error, texto = s.herramienta("consultar_disponibilidad", pista_id="P9", fecha="2026-12-01")
comprobar("una pista que no existe da un error de herramienta", es_error and "P9" in texto)

es_error, texto = s.herramienta("consultar_disponibilidad", pista_id="P2", fecha="1/12/2026")
comprobar("una fecha mal escrita se rechaza antes de llegar al dominio", es_error)

es_error, texto = s.herramienta("crear_reserva", socio_id="S2", pista_id="P1", inicio="2026-12-04T10:00")
comprobar("sin permiso, crear_reserva se niega aunque esté en el catálogo", es_error and "desactivada" in texto)
print("      mensaje:", texto)

desconocido = s.pedir("resources/list")
comprobar("un método no soportado devuelve el error -32601", desconocido["error"]["code"] == -32601)
s.cerrar()

# Sesión 2: una persona ha activado la escritura
s = Sesion(escritura=True)
s.iniciar()
es_error, texto = s.herramienta("crear_reserva", socio_id="S2", pista_id="P1", inicio="2026-12-04T10:00")
comprobar("con permiso, una reserva válida se crea", not es_error and json.loads(texto)["estado"] == "confirmada")
es_error, texto = s.herramienta("crear_reserva", socio_id="S2", pista_id="P2", inicio="2026-12-02T19:00")
comprobar("se rechaza una reserva que se solapa", es_error and "ya está reservada" in texto)
es_error, texto = s.herramienta("crear_reserva", socio_id="S1", pista_id="P3", inicio="2026-12-05T11:00")
comprobar("se rechaza la tercera reserva futura de un socio", es_error and "2 reservas" in texto)
es_error, texto = s.herramienta("crear_reserva", socio_id="S3", pista_id="P3", inicio="2026-12-05T11:00")
comprobar("se rechaza a un socio inactivo", es_error and "activos" in texto)
es_error, texto = s.herramienta("crear_reserva", socio_id="S2", pista_id="P3", inicio="2026-12-20T11:00")
comprobar("se rechaza reservar con más de 14 días", es_error and "14 días" in texto)
s.cerrar()

print(f"\n{'Todo correcto' if fallos == 0 else f'{fallos} comprobaciones fallidas'}")
sys.exit(1 if fallos else 0)
