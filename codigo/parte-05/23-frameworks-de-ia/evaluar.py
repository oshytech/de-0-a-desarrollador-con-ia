"""Evaluar la recuperación del RAG con un conjunto de referencia.

Cada pregunta lleva anotado a mano el fragmento que debería encontrarse.
Se mide en cuántas preguntas ese fragmento sale el primero y en cuántas
aparece entre los tres primeros, con y sin el filtro de documentos vigentes.

Uso: python evaluar.py
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))  # para importar rag también con python -I
from rag import Indice, trocear

sys.stdout.reconfigure(encoding="utf-8")

REFERENCIA = [
    ("¿Hasta cuándo puedo cancelar una reserva?", "normativa-2026#Cancelaciones"),
    ("¿Cuántos invitados puedo llevar?", "normativa-2026#Invitados"),
    ("¿Cuánto paga un invitado?", "tarifas-2026#Invitados"),
    ("¿Cuánto cuesta la hora de pádel?", "tarifas-2026#Pistas"),
    ("¿Cuándo me cobran la cuota?", "tarifas-2026#Cuota de socio"),
    ("¿Cuántas reservas puedo tener a la vez?", "normativa-2026#Límite por socio"),
    ("¿Cuántas pistas de pádel hay?", "instalaciones#Pistas"),
    ("¿A qué hora abre el bar el sábado?", "instalaciones#Bar"),
    ("¿Hasta qué hora se puede reservar?", "normativa-2026#Horario"),
    ("¿Puedo reservar para dentro de tres semanas?", "normativa-2026#Antelación"),
]


def evaluar(indice, solo_vigentes):
    primero, entre_tres, fallos = 0, 0, []
    for pregunta, esperada in REFERENCIA:
        citas = [f.cita for _, f in indice.recuperar(pregunta, k=3, solo_vigentes=solo_vigentes)]
        primero += bool(citas) and citas[0] == esperada
        entre_tres += esperada in citas
        if esperada not in citas:
            fallos.append((pregunta, citas))
    n = len(REFERENCIA)
    print(f"solo vigentes={solo_vigentes!s:<5}  primero: {primero}/{n}  entre los tres primeros: {entre_tres}/{n}")
    for pregunta, citas in fallos:
        print(f"    no encontrada: {pregunta}  -> {citas or 'nada por encima del umbral'}")


if __name__ == "__main__":
    indice = Indice(trocear())
    evaluar(indice, solo_vigentes=False)
    evaluar(indice, solo_vigentes=True)
