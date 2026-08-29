"""Pedir a un modelo multimodal el texto alternativo de una imagen.

La imagen viaja en el mismo mensaje que el texto, codificada en base64.
Si no pasas ninguna imagen, el script genera un PNG diminuto de prueba.

Sin LLM_API_KEY funciona en modo simulado. Con clave, el modelo
configurado tiene que aceptar imágenes.

Uso: python describir_imagen.py [ruta/a/imagen.png]
"""
import base64
import struct
import sys
import zlib
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))  # para importar cliente_llm también con python -I
from cliente_llm import completar, hay_modelo_real

sys.stdout.reconfigure(encoding="utf-8")

INSTRUCCIONES = (
    "Escribe el texto alternativo de esta imagen para la web del club. "
    "Una sola frase en español, de menos de 150 caracteres, que describa "
    "lo que se ve sin empezar por 'Imagen de'. Si hay texto en la imagen, inclúyelo."
)


def png_de_prueba(ancho=8, alto=8):
    """Un cuadrado verde en PNG, generado sin librerías externas."""
    fila = b"\x00" + bytes([40, 160, 70]) * ancho
    datos = zlib.compress(fila * alto)

    def bloque(tipo, contenido):
        crc = zlib.crc32(tipo + contenido) & 0xFFFFFFFF
        return struct.pack(">I", len(contenido)) + tipo + contenido + struct.pack(">I", crc)

    cabecera = struct.pack(">IIBBBBB", ancho, alto, 8, 2, 0, 0, 0)
    return b"\x89PNG\r\n\x1a\n" + bloque(b"IHDR", cabecera) + bloque(b"IDAT", datos) + bloque(b"IEND", b"")


def mensajes_con_imagen(imagen_png):
    codificada = base64.b64encode(imagen_png).decode("ascii")
    return [{
        "role": "user",
        "content": [
            {"type": "text", "text": INSTRUCCIONES},
            {"type": "image_url", "image_url": {"url": f"data:image/png;base64,{codificada}"}},
        ],
    }]


def simulador(mensajes):
    return "Pista de pádel vacía con césped artificial verde."


def validar_alt(texto):
    texto = texto.strip()
    if not texto or len(texto) >= 150 or texto.lower().startswith("imagen de"):
        return None
    return texto


if __name__ == "__main__":
    ruta = Path(sys.argv[1]) if len(sys.argv) > 1 else None
    imagen = ruta.read_bytes() if ruta else png_de_prueba()
    mensajes = mensajes_con_imagen(imagen)
    print("Modo:", "modelo real" if hay_modelo_real() else "simulado (sin LLM_API_KEY)")
    print(f"Imagen: {len(imagen)} bytes, {len(mensajes[0]['content'][1]['image_url']['url'])} caracteres en base64")
    alt = validar_alt(completar(mensajes, simulador))
    print("Texto alternativo:", alt if alt else "respuesta no válida, revisar a mano")
