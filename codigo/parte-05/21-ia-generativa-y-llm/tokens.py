"""Un tokenizador BPE de juguete.

Los tokenizadores reales (BPE y variantes) aprenden, a partir de muchísimo
texto, qué trozos de palabra aparecen juntos con frecuencia y los fusionan
en un solo token. Este hace lo mismo con un corpus de cuatro frases, para
ver la idea: las palabras frecuentes acaban en pocos tokens y las raras
se parten en trozos.

Uso: python tokens.py
"""
import sys
from collections import Counter

sys.stdout.reconfigure(encoding="utf-8")

CORPUS = """
reserva la pista de tenis para mañana y reserva la pista de pádel
las reservas se confirman al momento y las reservas canceladas se liberan
puedes reservar una hora y cancelar la reserva hasta dos horas antes
la pista de tenis y la pista de pádel se reservan por horas
"""

FIN = "_"  # marca de final de palabra: "pista_" es "pista" al final de una palabra


def palabras_como_simbolos(texto):
    """Cada palabra empieza partida en letras, con una marca de final."""
    return Counter(tuple(palabra) + (FIN,) for palabra in texto.split())


def par_mas_frecuente(vocabulario):
    pares = Counter()
    for simbolos, veces in vocabulario.items():
        for a, b in zip(simbolos, simbolos[1:]):
            pares[(a, b)] += veces
    return pares.most_common(1)[0] if pares else None


def fusionar(simbolos, par):
    resultado, i = [], 0
    while i < len(simbolos):
        if i + 1 < len(simbolos) and (simbolos[i], simbolos[i + 1]) == par:
            resultado.append(simbolos[i] + simbolos[i + 1])
            i += 2
        else:
            resultado.append(simbolos[i])
            i += 1
    return tuple(resultado)


def entrenar(texto, fusiones):
    vocabulario = palabras_como_simbolos(texto)
    reglas = []
    for _ in range(fusiones):
        mejor = par_mas_frecuente(vocabulario)
        if mejor is None or mejor[1] < 2:
            break
        par = mejor[0]
        reglas.append(par)
        vocabulario = Counter({fusionar(s, par): v for s, v in vocabulario.items()})
    return reglas


def tokenizar(palabra, reglas):
    simbolos = tuple(palabra) + (FIN,)
    for par in reglas:
        simbolos = fusionar(simbolos, par)
    return list(simbolos)


if __name__ == "__main__":
    reglas = entrenar(CORPUS, fusiones=40)
    print(f"Fusiones aprendidas: {len(reglas)}. Las diez primeras:")
    for a, b in reglas[:10]:
        print(f"  {a!r} + {b!r}")
    print()
    for palabra in ["reserva", "reservas", "reservar", "pista", "pádel", "cancelación", "frontón"]:
        trozos = tokenizar(palabra, reglas)
        print(f"{palabra:<12} -> {len(trozos)} tokens: {' | '.join(trozos)}")
