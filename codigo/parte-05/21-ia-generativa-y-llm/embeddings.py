"""Embeddings y similitud del coseno, con vectores hechos a mano.

Un modelo de embeddings real produce vectores de cientos o miles de
dimensiones que nadie ha diseñado a mano y que no tienen un significado
legible. Aquí cada vector tiene cuatro dimensiones con nombre para que se
pueda seguir el cálculo: [deporte, reserva, dinero, queja].

Uso: python embeddings.py
"""
import math
import sys

sys.stdout.reconfigure(encoding="utf-8")

VECTORES = {
    "pádel":       [0.9, 0.2, 0.0, 0.0],
    "tenis":       [0.9, 0.1, 0.0, 0.0],
    "raqueta":     [0.8, 0.0, 0.1, 0.0],
    "reservar":    [0.2, 0.9, 0.1, 0.0],
    "cancelar":    [0.1, 0.8, 0.2, 0.3],
    "cuota":       [0.1, 0.1, 0.9, 0.1],
    "reembolso":   [0.0, 0.3, 0.8, 0.4],
    "reclamación": [0.0, 0.2, 0.3, 0.9],
}


def coseno(a, b):
    """1 si apuntan en la misma dirección, 0 si no tienen nada que ver."""
    producto = sum(x * y for x, y in zip(a, b))
    norma_a = math.sqrt(sum(x * x for x in a))
    norma_b = math.sqrt(sum(y * y for y in b))
    return producto / (norma_a * norma_b)


def vecinos(palabra, n=3):
    origen = VECTORES[palabra]
    otras = [(otra, coseno(origen, v)) for otra, v in VECTORES.items() if otra != palabra]
    return sorted(otras, key=lambda par: par[1], reverse=True)[:n]


if __name__ == "__main__":
    for palabra in ["pádel", "cancelar", "reembolso"]:
        cercanas = ", ".join(f"{p} ({s:.2f})" for p, s in vecinos(palabra))
        print(f"{palabra:<10} -> {cercanas}")

    # Una frase se puede representar con la media de los vectores de sus palabras.
    frase = ["cancelar", "reembolso"]
    media = [sum(VECTORES[p][i] for p in frase) / len(frase) for i in range(4)]
    print()
    print(f"Frase {frase} -> más parecida a:")
    for palabra, vector in sorted(VECTORES.items(), key=lambda kv: -coseno(media, kv[1]))[:3]:
        print(f"  {palabra:<12} {coseno(media, vector):.2f}")
