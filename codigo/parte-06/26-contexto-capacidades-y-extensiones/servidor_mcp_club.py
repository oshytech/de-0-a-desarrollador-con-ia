"""Servidor MCP mínimo para el club deportivo, solo con la librería estándar.

Habla JSON-RPC 2.0 por stdio, un mensaje JSON por línea, que es el transporte
que usan los servidores MCP locales. En un proyecto real usarías el SDK oficial
del protocolo. Aquí está escrito a mano para que se vea qué viaja por el cable.

Herramientas que expone:
  listar_pistas             lectura
  consultar_disponibilidad  lectura
  reservas_de_socio         lectura
  crear_reserva             escritura, desactivada salvo CLUB_MCP_ESCRITURA=1

Reglas del servidor:
  - stdout es solo para el protocolo. Los registros van a stderr.
  - La escritura se controla aquí, no en lo que el modelo prometa hacer.
  - El fichero datos/club.json no se modifica nunca: las reservas nuevas
    viven en memoria mientras dura el proceso.
  - El reloj es fijo (CLUB_AHORA) para que el ejemplo sea reproducible.

Uso: lo lanza un agente (ver configuracion-mcp.json) o probar_servidor.py.
"""
import json
import os
import sys
from datetime import datetime, timedelta
from pathlib import Path

VERSIONES_PROTOCOLO = ["2025-06-18", "2025-03-26", "2024-11-05"]
DATOS = Path(__file__).parent / "datos" / "club.json"
FORMATO_FECHA = "%Y-%m-%d"
FORMATO_INICIO = "%Y-%m-%dT%H:%M"


def registrar(mensaje):
    print(f"[club-mcp] {mensaje}", file=sys.stderr, flush=True)


class ErrorDeHerramienta(Exception):
    """Error de entrada o de dominio: se devuelve al modelo como resultado con isError."""


# ---------------------------------------------------------------- dominio

class Club:
    def __init__(self, datos, ahora, escritura):
        self.socios = {s["id"]: s for s in datos["socios"]}
        self.pistas = {p["id"]: p for p in datos["pistas"]}
        self.reservas = list(datos["reservas"])
        self.ahora = ahora
        self.escritura = escritura

    def _pista(self, pista_id):
        if pista_id not in self.pistas:
            raise ErrorDeHerramienta(f"No existe la pista {pista_id}.")
        return self.pistas[pista_id]

    def _confirmadas(self):
        return [r for r in self.reservas if r["estado"] == "confirmada"]

    def listar_pistas(self, _args):
        return list(self.pistas.values())

    def consultar_disponibilidad(self, args):
        pista = self._pista(texto(args, "pista_id"))
        dia = fecha(texto(args, "fecha"))
        ocupadas = {r["inicio"] for r in self._confirmadas() if r["pista"] == pista["id"]}
        libres = []
        for hora in range(8, 22):  # la última reserva empieza a las 21:00
            inicio = dia.replace(hour=hora).strftime(FORMATO_INICIO)
            if inicio not in ocupadas:
                libres.append(f"{hora:02d}:00")
        return {"pista": pista["nombre"], "fecha": dia.strftime(FORMATO_FECHA), "horas_libres": libres}

    def reservas_de_socio(self, args):
        socio_id = texto(args, "socio_id")
        if socio_id not in self.socios:
            raise ErrorDeHerramienta(f"No existe el socio {socio_id}.")
        return [r for r in self.reservas if r["socio"] == socio_id]

    def crear_reserva(self, args):
        if not self.escritura:
            raise ErrorDeHerramienta(
                "La escritura está desactivada en este servidor. "
                "Pide a una persona que active CLUB_MCP_ESCRITURA=1 si de verdad hace falta.")
        socio_id = texto(args, "socio_id")
        pista = self._pista(texto(args, "pista_id"))
        inicio = momento(texto(args, "inicio"))
        socio = self.socios.get(socio_id)
        if socio is None:
            raise ErrorDeHerramienta(f"No existe el socio {socio_id}.")
        if not socio["activo"]:
            raise ErrorDeHerramienta("Solo los socios activos pueden reservar.")
        if inicio.minute != 0 or not 8 <= inicio.hour <= 21:
            raise ErrorDeHerramienta("Las reservas empiezan en punto, entre las 08:00 y las 21:00.")
        if inicio <= self.ahora:
            raise ErrorDeHerramienta("No se puede reservar en el pasado.")
        if inicio > self.ahora + timedelta(days=14):
            raise ErrorDeHerramienta("No se puede reservar con más de 14 días de antelación.")
        texto_inicio = inicio.strftime(FORMATO_INICIO)
        if any(r["pista"] == pista["id"] and r["inicio"] == texto_inicio for r in self._confirmadas()):
            raise ErrorDeHerramienta(f"La pista {pista['nombre']} ya está reservada a esa hora.")
        futuras = [r for r in self._confirmadas()
                   if r["socio"] == socio_id and momento(r["inicio"]) > self.ahora]
        if len(futuras) >= 2:
            raise ErrorDeHerramienta("El socio ya tiene 2 reservas confirmadas futuras.")
        reserva = {"id": f"R{len(self.reservas) + 1}", "socio": socio_id, "pista": pista["id"],
                   "inicio": texto_inicio, "estado": "confirmada"}
        self.reservas.append(reserva)
        registrar(f"reserva creada en memoria: {reserva['id']}")
        return reserva


def texto(args, nombre):
    valor = args.get(nombre)
    if not isinstance(valor, str) or not valor.strip():
        raise ErrorDeHerramienta(f"Falta el campo {nombre} o no es un texto.")
    return valor.strip()


def fecha(valor):
    try:
        return datetime.strptime(valor, FORMATO_FECHA)
    except ValueError:
        raise ErrorDeHerramienta(f"La fecha {valor!r} no tiene el formato AAAA-MM-DD.") from None


def momento(valor):
    try:
        return datetime.strptime(valor, FORMATO_INICIO)
    except ValueError:
        raise ErrorDeHerramienta(f"El inicio {valor!r} no tiene el formato AAAA-MM-DDTHH:MM.") from None


# ------------------------------------------------------------ catálogo MCP

def esquema(propiedades):
    return {"type": "object",
            "properties": {n: {"type": "string", "description": d} for n, d in propiedades.items()},
            "required": list(propiedades), "additionalProperties": False}


HERRAMIENTAS = [
    {"name": "listar_pistas",
     "description": "Lista las pistas del club con su id, nombre y deporte.",
     "inputSchema": esquema({}),
     "annotations": {"readOnlyHint": True}},
    {"name": "consultar_disponibilidad",
     "description": "Devuelve las horas libres de una pista en un día. No reserva nada.",
     "inputSchema": esquema({"pista_id": "Id de la pista, por ejemplo P2",
                             "fecha": "Día en formato AAAA-MM-DD"}),
     "annotations": {"readOnlyHint": True}},
    {"name": "reservas_de_socio",
     "description": "Lista las reservas de un socio, confirmadas y canceladas.",
     "inputSchema": esquema({"socio_id": "Id del socio, por ejemplo S1"}),
     "annotations": {"readOnlyHint": True}},
    {"name": "crear_reserva",
     "description": ("Crea una reserva de una hora si cumple las reglas del club. "
                     "Puede estar desactivada: en ese caso devuelve un error y no hay que insistir."),
     "inputSchema": esquema({"socio_id": "Id del socio",
                             "pista_id": "Id de la pista",
                             "inicio": "Inicio en formato AAAA-MM-DDTHH:MM, en punto"}),
     "annotations": {"readOnlyHint": False, "destructiveHint": False}},
]


# ------------------------------------------------------------- JSON-RPC

def respuesta(id_, resultado):
    return {"jsonrpc": "2.0", "id": id_, "result": resultado}


def error(id_, codigo, mensaje):
    return {"jsonrpc": "2.0", "id": id_, "error": {"code": codigo, "message": mensaje}}


def atender(club, mensaje):
    """Procesa un mensaje y devuelve la respuesta, o None si era una notificación."""
    metodo = mensaje.get("method")
    id_ = mensaje.get("id")
    es_notificacion = "id" not in mensaje
    params = mensaje.get("params") or {}

    if metodo == "initialize":
        pedida = params.get("protocolVersion")
        version = pedida if pedida in VERSIONES_PROTOCOLO else VERSIONES_PROTOCOLO[0]
        return respuesta(id_, {
            "protocolVersion": version,
            "capabilities": {"tools": {}},
            "serverInfo": {"name": "club-reservas", "version": "0.1.0"},
            "instructions": "Datos ficticios del club. Consulta antes de proponer una reserva.",
        })
    if es_notificacion:
        registrar(f"notificación recibida: {metodo}")
        return None
    if metodo == "ping":
        return respuesta(id_, {})
    if metodo == "tools/list":
        return respuesta(id_, {"tools": HERRAMIENTAS})
    if metodo == "tools/call":
        nombre = params.get("name")
        if nombre not in {h["name"] for h in HERRAMIENTAS}:
            return error(id_, -32602, f"Herramienta desconocida: {nombre}")
        try:
            resultado = getattr(club, nombre)(params.get("arguments") or {})
            contenido = json.dumps(resultado, ensure_ascii=False, indent=2)
            return respuesta(id_, {"content": [{"type": "text", "text": contenido}], "isError": False})
        except ErrorDeHerramienta as e:
            return respuesta(id_, {"content": [{"type": "text", "text": str(e)}], "isError": True})
    return error(id_, -32601, f"Método no soportado: {metodo}")


def main():
    datos = json.loads(DATOS.read_text(encoding="utf-8"))
    ahora = datetime.strptime(os.environ.get("CLUB_AHORA", "2026-12-01T10:00"), FORMATO_INICIO)
    escritura = os.environ.get("CLUB_MCP_ESCRITURA") == "1"
    club = Club(datos, ahora, escritura)
    registrar(f"arrancado (escritura {'activada' if escritura else 'desactivada'})")

    entrada = sys.stdin.buffer
    salida = sys.stdout.buffer
    for linea in entrada:
        if not linea.strip():
            continue
        try:
            mensaje = json.loads(linea)
        except json.JSONDecodeError:
            contestacion = error(None, -32700, "El mensaje no es JSON válido.")
        else:
            contestacion = atender(club, mensaje)
        if contestacion is not None:
            salida.write(json.dumps(contestacion, ensure_ascii=False).encode("utf-8") + b"\n")
            salida.flush()
    registrar("entrada cerrada, fin")


if __name__ == "__main__":
    main()
