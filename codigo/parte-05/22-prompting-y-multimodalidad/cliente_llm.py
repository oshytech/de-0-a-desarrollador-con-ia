"""Cliente mínimo para cualquier API compatible con el formato de chat de OpenAI.

Muchos proveedores, y también servidores locales como Ollama, aceptan este
formato. Lee la configuración de variables de entorno:

  LLM_API_KEY   clave del proveedor (si no existe, se usa el modo simulado)
  LLM_BASE_URL  por ejemplo https://api.proveedor.com/v1 o http://localhost:11434/v1
  LLM_MODEL     nombre del modelo

Sin clave no se hace ninguna llamada de red: responde la función `simulador`
que le pases, para poder ejecutar los ejemplos sin cuenta ni coste.
"""
import json
import os
import urllib.request


def hay_modelo_real():
    return bool(os.environ.get("LLM_API_KEY"))


def completar(mensajes, simulador, temperatura=0.0):
    """Envía los mensajes al modelo y devuelve el texto de la respuesta."""
    if not hay_modelo_real():
        return simulador(mensajes)

    base = os.environ.get("LLM_BASE_URL", "http://localhost:11434/v1").rstrip("/")
    cuerpo = {
        "model": os.environ.get("LLM_MODEL", "llama3.1:8b"),
        "messages": mensajes,
        "temperature": temperatura,
    }
    peticion = urllib.request.Request(
        f"{base}/chat/completions",
        data=json.dumps(cuerpo).encode("utf-8"),
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {os.environ['LLM_API_KEY']}",
        },
    )
    with urllib.request.urlopen(peticion, timeout=60) as respuesta:
        datos = json.load(respuesta)
    return datos["choices"][0]["message"]["content"]
