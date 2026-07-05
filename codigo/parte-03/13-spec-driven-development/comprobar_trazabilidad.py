"""Comprueba que cada regla de una spec tiene al menos un escenario y una tarea.

Uso: python -I comprobar_trazabilidad.py specs/reservar-pista
Devuelve 1 si hay reglas sin cubrir o referencias a reglas que no existen.
"""
import re
import sys
from pathlib import Path


def ids_de_reglas(texto):
    return set(re.findall(r"^- \*\*([A-Z]\d+)\.\*\*", texto, flags=re.M))


def ids_de_escenarios(texto):
    return re.findall(r"^\s*@([A-Z]\d+)\s*$", texto, flags=re.M)


def ids_de_tareas(texto):
    cubre = re.findall(r"Cubre: ([A-Z0-9, ]+)\.", texto)
    return [i.strip() for linea in cubre for i in linea.split(",")]


def main(carpeta):
    carpeta = Path(carpeta)
    reglas = ids_de_reglas((carpeta / "requisitos.md").read_text(encoding="utf-8"))
    escenarios = [i for f in carpeta.glob("*.feature") for i in ids_de_escenarios(f.read_text(encoding="utf-8"))]
    tareas = ids_de_tareas((carpeta / "tareas.md").read_text(encoding="utf-8"))

    problemas = 0
    print(f"{'Regla':<6} {'Escenarios':>10} {'Tareas':>7}")
    for regla in sorted(reglas):
        n_esc, n_tar = escenarios.count(regla), tareas.count(regla)
        aviso = "  <- sin cubrir" if n_esc == 0 or n_tar == 0 else ""
        problemas += bool(aviso)
        print(f"{regla:<6} {n_esc:>10} {n_tar:>7}{aviso}")
    for huerfano in sorted(set(escenarios + tareas) - reglas):
        print(f"{huerfano}: aparece en escenarios o tareas, pero no en la spec")
        problemas += 1
    return 1 if problemas else 0


if __name__ == "__main__":
    sys.stdout.reconfigure(encoding="utf-8")
    sys.exit(main(sys.argv[1] if len(sys.argv) > 1 else "specs/reservar-pista"))
