"""Cuentas: reparto de gastos compartidos desde la terminal.

Ejemplos:
    python main.py anadir "Compra del súper" 63,40 --paga Ana --entre Ana Bruno Carla
    python main.py gastos
    python main.py saldos
    python main.py liquidar
"""

import argparse
import sys

import almacen
import reparto

FICHERO_POR_DEFECTO = "cuentas.json"


def crear_parser():
    parser = argparse.ArgumentParser(description="Reparto de gastos compartidos")
    parser.add_argument(
        "--fichero", default=FICHERO_POR_DEFECTO, help="dónde se guardan los gastos"
    )
    ordenes = parser.add_subparsers(dest="orden", required=True)

    anadir = ordenes.add_parser("anadir", help="registra un gasto")
    anadir.add_argument("concepto")
    anadir.add_argument("importe", help="en euros, por ejemplo 12,50")
    anadir.add_argument("--paga", required=True, help="quién ha pagado")
    anadir.add_argument(
        "--entre", nargs="+", required=True, help="entre quiénes se reparte"
    )

    ordenes.add_parser("gastos", help="lista los gastos registrados")
    ordenes.add_parser("saldos", help="muestra cuánto debe o le deben a cada persona")
    ordenes.add_parser("liquidar", help="propone los pagos para quedar en paz")
    return parser


def orden_anadir(args, gastos):
    gasto = {
        "concepto": args.concepto,
        "pagado_por": args.paga,
        "importe": reparto.a_centimos(args.importe),
        "entre": args.entre,
    }
    reparto.repartir(gasto["importe"], gasto["entre"])  # valida antes de guardar
    gastos.append(gasto)
    almacen.guardar(args.fichero, gastos)
    importe = reparto.formatear(gasto["importe"])
    print(f"Añadido: {gasto['concepto']} ({importe}), pagado por {gasto['pagado_por']}")


def orden_gastos(gastos):
    if not gastos:
        print("Todavía no hay gastos")
    for numero, gasto in enumerate(gastos, start=1):
        personas = ", ".join(gasto["entre"])
        print(
            f"{numero}. {gasto['concepto']}: {reparto.formatear(gasto['importe'])} "
            f"(paga {gasto['pagado_por']}, entre {personas})"
        )


def orden_saldos(gastos):
    if not gastos:
        print("Todavía no hay gastos")
    for persona, saldo in sorted(reparto.calcular_saldos(gastos).items()):
        print(f"{persona:>8}: {reparto.formatear(saldo):>10}")


def orden_liquidar(gastos):
    pagos = reparto.liquidar(reparto.calcular_saldos(gastos))
    if not pagos:
        print("Estáis en paz")
    for deudor, acreedor, cantidad in pagos:
        print(f"{deudor} paga {reparto.formatear(cantidad)} a {acreedor}")


def main(argumentos=None):
    args = crear_parser().parse_args(argumentos)
    try:
        gastos = almacen.cargar(args.fichero)
        if args.orden == "anadir":
            orden_anadir(args, gastos)
        elif args.orden == "gastos":
            orden_gastos(gastos)
        elif args.orden == "saldos":
            orden_saldos(gastos)
        elif args.orden == "liquidar":
            orden_liquidar(gastos)
    except (ValueError, almacen.FicheroDanado) as error:
        print(f"Error: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
