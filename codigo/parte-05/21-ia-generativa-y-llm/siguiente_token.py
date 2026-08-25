"""Cómo elige un modelo de lenguaje la siguiente palabra.

Las puntuaciones (logits) son inventadas para el ejemplo: un modelo real
las calcula para decenas de miles de tokens posibles. Lo que sí es real es
la mecánica: softmax con temperatura, top-k, top-p y sorteo.

Uso: python siguiente_token.py
"""
import math
import random
import sys

sys.stdout.reconfigure(encoding="utf-8")

CONTEXTO = "Quiero reservar la pista de"

# Puntuaciones de ejemplo para la palabra que sigue al contexto.
LOGITS = {
    "pádel": 3.1,
    "tenis": 2.8,
    "las": 1.2,
    "mañana": 0.9,
    "siempre": 0.3,
    "baloncesto": -0.5,
    "patinaje": -1.4,
}


def softmax(logits, temperatura):
    """Convierte puntuaciones en probabilidades. La temperatura las aplana o las afila."""
    escalados = {t: v / temperatura for t, v in logits.items()}
    maximo = max(escalados.values())
    exps = {t: math.exp(v - maximo) for t, v in escalados.items()}
    total = sum(exps.values())
    return {t: e / total for t, e in exps.items()}


def filtrar_top_k(probs, k):
    """Se queda con los k candidatos más probables y reparte de nuevo la probabilidad."""
    mejores = sorted(probs.items(), key=lambda par: par[1], reverse=True)[:k]
    total = sum(p for _, p in mejores)
    return {t: p / total for t, p in mejores}


def filtrar_top_p(probs, p_max):
    """Se queda con los candidatos más probables hasta acumular p_max de probabilidad."""
    elegidos, acumulado = [], 0.0
    for token, p in sorted(probs.items(), key=lambda par: par[1], reverse=True):
        elegidos.append((token, p))
        acumulado += p
        if acumulado >= p_max:
            break
    total = sum(p for _, p in elegidos)
    return {t: p / total for t, p in elegidos}


def sortear(probs, rng):
    tokens = list(probs)
    return rng.choices(tokens, weights=[probs[t] for t in tokens])[0]


def mostrar(titulo, probs):
    print(titulo)
    for token, p in sorted(probs.items(), key=lambda par: par[1], reverse=True):
        print(f"  {token:<11} {p:6.1%}  {'#' * round(p * 40)}")
    print()


if __name__ == "__main__":
    print(f'Contexto: "{CONTEXTO} ..."\n')
    for t in (0.2, 1.0, 1.5):
        mostrar(f"Temperatura {t}", softmax(LOGITS, t))

    base = softmax(LOGITS, 1.0)
    mostrar("Temperatura 1.0 con top-k = 3", filtrar_top_k(base, 3))
    mostrar("Temperatura 1.0 con top-p = 0.9", filtrar_top_p(base, 0.9))

    rng = random.Random(42)  # semilla fija para que el resultado sea reproducible
    for t in (0.2, 1.5):
        probs = softmax(LOGITS, t)
        conteo = {}
        for _ in range(1000):
            elegido = sortear(probs, rng)
            conteo[elegido] = conteo.get(elegido, 0) + 1
        resumen = ", ".join(f"{tok} {n}" for tok, n in sorted(conteo.items(), key=lambda par: -par[1]))
        print(f"1000 sorteos con temperatura {t}: {resumen}")
