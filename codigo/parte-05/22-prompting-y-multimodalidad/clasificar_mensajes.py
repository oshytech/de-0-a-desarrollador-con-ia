"""Clasificar los mensajes que los socios envían al club, con dos versiones de prompt.

Compara un prompt vago con uno diseñado (mensaje de sistema, categorías
definidas, ejemplos, delimitadores y salida JSON) sobre un pequeño conjunto
de mensajes etiquetados a mano, y valida cada respuesta antes de usarla.

Sin LLM_API_KEY funciona en modo simulado: las respuestas salen de un
clasificador por palabras clave que imita el formato de un modelo. Sirve
para ejecutar el código de principio a fin, no para medir ningún modelo.

Uso: python clasificar_mensajes.py
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))  # para importar cliente_llm también con python -I
from cliente_llm import completar, hay_modelo_real

sys.stdout.reconfigure(encoding="utf-8")

CATEGORIAS = {"reserva", "cancelacion", "incidencia", "cuota", "otro"}

# Conjunto de prueba etiquetado a mano: (mensaje, categoría correcta).
ETIQUETADOS = [
    ("¿Queda alguna pista de pádel libre el viernes a las 19:00?", "reserva"),
    ("Al final no puedo ir mañana, anulad mi reserva de tenis.", "cancelacion"),
    ("La red de la pista 3 está rota y se ha soltado un poste.", "incidencia"),
    ("Me habéis cobrado la cuota de octubre dos veces.", "cuota"),
    ("No puedo ir el sábado, ¿me devolvéis el dinero de la reserva?", "cancelacion"),
    ("¿A qué hora abre el bar del club los domingos?", "otro"),
    ("Quiero reservar tenis para el martes que viene por la mañana.", "reserva"),
    ("En la pista 1 no funcionan los focos y no se ve nada.", "incidencia"),
]


def prompt_vago(mensaje):
    return [{"role": "user", "content": f"Clasifica este mensaje: {mensaje}"}]


SISTEMA = """Eres el asistente que ordena el buzón de un club deportivo de tenis y pádel.
Clasifica el mensaje del socio en UNA de estas categorías:
- reserva: quiere reservar o pregunta por disponibilidad.
- cancelacion: quiere anular una reserva, aunque pregunte por el dinero.
- incidencia: algo de las instalaciones está roto o no funciona.
- cuota: pagos, recibos o cobros de la cuota de socio.
- otro: cualquier otra cosa.
Marca urgente=true solo si hay riesgo para la seguridad de alguien.
Responde solo con JSON, sin texto alrededor, con esta forma:
{"categoria": "...", "urgente": false}"""

EJEMPLOS = [
    ("Se ha caído la valla del fondo de la pista 2.", '{"categoria": "incidencia", "urgente": true}'),
    ("¿Puedo pagar la cuota con tarjeta?", '{"categoria": "cuota", "urgente": false}'),
]


def prompt_disenado(mensaje):
    mensajes = [{"role": "system", "content": SISTEMA}]
    for ejemplo, respuesta in EJEMPLOS:
        mensajes.append({"role": "user", "content": f"<mensaje>{ejemplo}</mensaje>"})
        mensajes.append({"role": "assistant", "content": respuesta})
    mensajes.append({"role": "user", "content": f"<mensaje>{mensaje}</mensaje>"})
    return mensajes


def validar(texto):
    """Devuelve el dict si la respuesta es utilizable. Si no, None."""
    limpio = texto.strip().removeprefix("```json").removesuffix("```").strip()
    try:
        datos = json.loads(limpio)
    except json.JSONDecodeError:
        return None
    if not isinstance(datos, dict) or datos.get("categoria") not in CATEGORIAS:
        return None
    if not isinstance(datos.get("urgente"), bool):
        return None
    return datos


def simulador(mensajes):
    """Imita a un modelo: en prosa si nadie le pide formato, en JSON si se lo piden."""
    texto = mensajes[-1]["content"].lower()
    reglas = [
        ("incidencia", ("rota", "roto", "no funciona", "se ha caído")),
        ("cuota", ("cuota", "cobrado", "dinero", "pagar")),
        ("cancelacion", ("anula", "no puedo ir")),
        ("reserva", ("reservar", "libre")),
    ]
    categoria = next((c for c, claves in reglas if any(k in texto for k in claves)), "otro")
    pide_json = any("JSON" in m["content"] for m in mensajes)
    if not pide_json:
        return f"Este mensaje parece estar relacionado con: {categoria}."
    return "```json\n" + json.dumps({"categoria": categoria, "urgente": False}) + "\n```"


def evaluar(nombre, construir_prompt):
    aciertos, invalidas = 0, 0
    for mensaje, esperada in ETIQUETADOS:
        respuesta = completar(construir_prompt(mensaje), simulador)
        datos = validar(respuesta)
        if datos is None:
            invalidas += 1
        elif datos["categoria"] == esperada:
            aciertos += 1
        else:
            print(f"  fallo [{nombre}]: {mensaje!r} -> {datos['categoria']} (esperada: {esperada})")
    total = len(ETIQUETADOS)
    print(f"{nombre:<10} aciertos {aciertos}/{total}, respuestas no válidas {invalidas}/{total}")


if __name__ == "__main__":
    print("Modo:", "modelo real" if hay_modelo_real() else "simulado (sin LLM_API_KEY)")
    evaluar("vago", prompt_vago)
    evaluar("diseñado", prompt_disenado)
