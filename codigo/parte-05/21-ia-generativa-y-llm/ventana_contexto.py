"""Por qué un asistente "olvida" el principio de una conversación larga.

Cada petición a un modelo incluye la conversación entera, y todo tiene que
caber en la ventana de contexto. Cuando no cabe, la aplicación descarta
algo, normalmente los mensajes más antiguos. Aquí el límite es diminuto
para que se vea, y los tokens se estiman con una regla aproximada
(unos 4 caracteres por token), no con un tokenizador real.

Uso: python ventana_contexto.py
"""
import math
import sys

sys.stdout.reconfigure(encoding="utf-8")

LIMITE_TOKENS = 45


def estimar_tokens(texto):
    return math.ceil(len(texto) / 4)


def recortar(mensajes, limite):
    """Conserva siempre el mensaje de sistema y añade los más recientes que quepan."""
    sistema, resto = mensajes[0], mensajes[1:]
    usados = estimar_tokens(sistema["texto"])
    conservados = []
    for mensaje in reversed(resto):
        coste = estimar_tokens(mensaje["texto"])
        if usados + coste > limite:
            break
        conservados.insert(0, mensaje)
        usados += coste
    return [sistema] + conservados, usados


if __name__ == "__main__":
    conversacion = [
        {"rol": "sistema", "texto": "Eres el asistente de reservas del club."},
        {"rol": "usuario", "texto": "Me llamo Laura y soy socia desde 2019."},
        {"rol": "asistente", "texto": "Hola, Laura. ¿En qué te ayudo?"},
        {"rol": "usuario", "texto": "Quiero pádel el jueves por la tarde."},
        {"rol": "asistente", "texto": "Tengo la pista 2 a las 18:00 y a las 19:00."},
        {"rol": "usuario", "texto": "La de las 19:00. ¿Cómo me llamo?"},
    ]
    enviados, usados = recortar(conversacion, LIMITE_TOKENS)
    print(f"Límite: {LIMITE_TOKENS} tokens. Usados: {usados}.\n")
    for mensaje in conversacion:
        marca = "enviado " if mensaje in enviados else "FUERA   "
        print(f"{marca} [{mensaje['rol']}] {mensaje['texto']}")
