"""Lógica del reparto de gastos. Funciones puras: no leen ficheros ni imprimen."""


class ImporteInvalido(ValueError):
    """El texto no representa un importe positivo con, como mucho, dos decimales."""


def a_centimos(texto):
    """Convierte "12,5", "12.50" o "12" en céntimos (1250, 1250, 1200)."""
    limpio = texto.strip().replace(",", ".")
    euros, _, decimales = limpio.partition(".")
    if (
        not euros.isdigit()
        or (decimales and not decimales.isdigit())
        or len(decimales) > 2
    ):
        raise ImporteInvalido(f"Importe no válido: {texto!r}")
    centimos = int(euros) * 100 + int(decimales.ljust(2, "0"))
    if centimos == 0:
        raise ImporteInvalido("El importe tiene que ser mayor que cero")
    return centimos


def formatear(centimos):
    """Convierte 1250 en "12,50 €" y -320 en "-3,20 €"."""
    signo = "-" if centimos < 0 else ""
    centimos = abs(centimos)
    return f"{signo}{centimos // 100},{centimos % 100:02d} €"


def repartir(importe, personas):
    """Reparte un importe en céntimos entre personas sin perder ni inventar céntimos.

    Los céntimos que sobran de la división se asignan, de uno en uno,
    a las primeras personas de la lista.
    """
    if not personas:
        raise ValueError("Un gasto tiene que repartirse entre al menos una persona")
    if len(set(personas)) != len(personas):
        raise ValueError(f"Hay personas repetidas en el reparto: {', '.join(personas)}")
    cuota, sobrante = divmod(importe, len(personas))
    return {
        persona: cuota + (1 if posicion < sobrante else 0)
        for posicion, persona in enumerate(personas)
    }


def calcular_saldos(gastos):
    """Saldo de cada persona en céntimos: positivo si le deben, negativo si debe."""
    saldos = {}
    for gasto in gastos:
        pagador = gasto["pagado_por"]
        saldos[pagador] = saldos.get(pagador, 0) + gasto["importe"]
        for persona, parte in repartir(gasto["importe"], gasto["entre"]).items():
            saldos[persona] = saldos.get(persona, 0) - parte
    return saldos


def liquidar(saldos):
    """Propone pagos (deudor, acreedor, céntimos) que dejan todos los saldos a cero.

    En cada paso, quien más debe le paga a quien más tiene que recibir.
    """
    if sum(saldos.values()) != 0:
        raise ValueError("Los saldos no cuadran: deberían sumar cero")
    pendientes = {persona: saldo for persona, saldo in saldos.items() if saldo != 0}
    pagos = []
    while pendientes:
        deudor = min(pendientes, key=pendientes.get)
        acreedor = max(pendientes, key=pendientes.get)
        cantidad = min(-pendientes[deudor], pendientes[acreedor])
        pagos.append((deudor, acreedor, cantidad))
        pendientes[deudor] += cantidad
        pendientes[acreedor] -= cantidad
        pendientes = {p: s for p, s in pendientes.items() if s != 0}
    return pagos
