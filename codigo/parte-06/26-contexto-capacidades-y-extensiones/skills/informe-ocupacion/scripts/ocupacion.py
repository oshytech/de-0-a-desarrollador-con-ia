"""Calcula la ocupación de cada pista en un periodo a partir del fichero de datos del club.

Uso: python -I ocupacion.py <datos.json> <desde AAAA-MM-DD> <hasta AAAA-MM-DD>
Cada día tiene 14 horas reservables (de 08:00 a 21:00 de inicio).
"""
import json
import sys
from collections import Counter
from datetime import date, datetime

sys.stdout.reconfigure(encoding="utf-8")
HORAS_POR_DIA = 14


def ocupacion(datos, desde, hasta):
    dias = (hasta - desde).days + 1
    if dias <= 0:
        raise ValueError("El periodo está vacío: 'hasta' es anterior a 'desde'.")
    confirmadas = [r for r in datos["reservas"] if r["estado"] == "confirmada"
                   and desde <= datetime.fromisoformat(r["inicio"]).date() <= hasta]
    por_pista = Counter(r["pista"] for r in confirmadas)
    disponibles = dias * HORAS_POR_DIA
    filas = [{"pista": p["nombre"], "reservadas": por_pista[p["id"]], "disponibles": disponibles,
              "porcentaje": round(100 * por_pista[p["id"]] / disponibles, 1)}
             for p in datos["pistas"]]
    franjas = Counter(r["inicio"][11:16] for r in confirmadas).most_common(3)
    return {"dias": dias, "pistas": filas, "franjas_mas_pedidas": franjas, "total": len(confirmadas)}


if __name__ == "__main__":
    if len(sys.argv) != 4:
        sys.exit(__doc__)
    fichero, desde, hasta = sys.argv[1], date.fromisoformat(sys.argv[2]), date.fromisoformat(sys.argv[3])
    with open(fichero, encoding="utf-8") as f:
        print(json.dumps(ocupacion(json.load(f), desde, hasta), ensure_ascii=False, indent=2))
