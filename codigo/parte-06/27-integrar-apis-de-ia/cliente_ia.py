"""Cliente para APIs de modelos compatibles con el formato de chat más extendido.

Es el cliente mínimo del capítulo 22 preparado para un programa de verdad:
tiempo máximo de espera, reintentos limitados, motivo de fin, consumo de
tokens y latencia de cada llamada.

Configuración por variables de entorno (ver .env.example):
  LLM_API_KEY   clave del proveedor. Sin clave ni URL, modo simulado.
  LLM_BASE_URL  por ejemplo https://api.proveedor.com/v1, o un servidor local
                compatible como http://localhost:11434/v1 (sin clave).
  LLM_MODEL     identificador exacto del modelo, copiado del catálogo del proveedor.

En modo simulado no sale nada a la red: responde la función `simulador` que
le pases. Sirve para ejecutar y probar el programa sin cuenta ni coste.
"""
import json
import os
import time
import urllib.error
import urllib.request
from dataclasses import dataclass

REINTENTABLES = {429, 500, 502, 503, 504}


class ErrorIA(Exception):
    pass


@dataclass
class Respuesta:
    texto: str
    motivo_fin: str
    tokens_entrada: int
    tokens_salida: int
    segundos: float
    simulada: bool


def modo_real():
    return bool(os.environ.get("LLM_API_KEY") or os.environ.get("LLM_BASE_URL"))


def preguntar(sistema, usuario, simulador, max_tokens=500, temperatura=0.0,
              reintentos=2, espera=1.0, timeout=30):
    """Envía un mensaje de sistema y uno de usuario y devuelve una Respuesta."""
    inicio = time.monotonic()
    if not modo_real():
        # En modo simulado los tokens son una estimación: unos cuatro caracteres por token.
        texto = simulador(sistema, usuario)
        return Respuesta(texto, "stop", len(sistema + usuario) // 4, len(texto) // 4,
                         time.monotonic() - inicio, simulada=True)

    modelo = os.environ.get("LLM_MODEL")
    if not modelo:
        raise ErrorIA("Falta LLM_MODEL. Copia el identificador exacto del catálogo del proveedor.")
    base = os.environ.get("LLM_BASE_URL", "").rstrip("/")
    if not base:
        raise ErrorIA("Falta LLM_BASE_URL.")
    cabeceras = {"Content-Type": "application/json"}
    if os.environ.get("LLM_API_KEY"):
        cabeceras["Authorization"] = f"Bearer {os.environ['LLM_API_KEY']}"
    cuerpo = json.dumps({
        "model": modelo,
        "messages": [{"role": "system", "content": sistema},
                     {"role": "user", "content": usuario}],
        "max_tokens": max_tokens,
        "temperature": temperatura,
    }).encode("utf-8")

    for intento in range(reintentos + 1):
        peticion = urllib.request.Request(f"{base}/chat/completions", data=cuerpo, headers=cabeceras)
        try:
            with urllib.request.urlopen(peticion, timeout=timeout) as r:
                datos = json.load(r)
            break
        except urllib.error.HTTPError as e:
            detalle = e.read().decode("utf-8", errors="replace")[:300]
            if e.code in REINTENTABLES and intento < reintentos:
                time.sleep(espera * 2 ** intento)
                continue
            raise ErrorIA(f"HTTP {e.code} del proveedor: {detalle}") from None
        except (urllib.error.URLError, TimeoutError) as e:
            if intento < reintentos:
                time.sleep(espera * 2 ** intento)
                continue
            raise ErrorIA(f"No se pudo contactar con el proveedor: {e}") from None

    eleccion = datos["choices"][0]
    uso = datos.get("usage", {})
    return Respuesta(
        texto=eleccion["message"].get("content") or "",
        motivo_fin=eleccion.get("finish_reason", "desconocido"),
        tokens_entrada=uso.get("prompt_tokens", 0),
        tokens_salida=uso.get("completion_tokens", 0),
        segundos=time.monotonic() - inicio,
        simulada=False,
    )
