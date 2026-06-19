"""Pide un JSON a un servidor HTTP y lo convierte en un diccionario de Python.

Antes, en otra terminal: cd servidor && python -m http.server 8000
"""
import json
import urllib.error
import urllib.request

BASE = "http://localhost:8000"


def pedir_json(ruta):
    peticion = urllib.request.Request(BASE + ruta, headers={"User-Agent": "cliente-tiempo/1.0"})
    with urllib.request.urlopen(peticion, timeout=5) as respuesta:
        print(f"{respuesta.status} {respuesta.reason} ({respuesta.headers['Content-Type']})")
        return json.loads(respuesta.read().decode("utf-8"))


for ruta in ["/tiempo.json", "/no-existe.json"]:
    try:
        datos = pedir_json(ruta)
        print(f"{datos['ciudad']}: {datos['temperatura']} grados, lluvia: {datos['lluvia']}")
    except urllib.error.HTTPError as error:
        print(f"El servidor respondió {error.code}: {error.reason}")
    except urllib.error.URLError as error:
        print(f"No se pudo conectar: {error.reason}")
