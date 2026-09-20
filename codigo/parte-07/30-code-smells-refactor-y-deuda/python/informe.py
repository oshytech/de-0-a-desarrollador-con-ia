"""Informe diario de ocupación de pistas."""

HORAS_RESERVABLES_POR_DIA = 14  # de 08:00 a 21:00, una reserva por hora
UMBRAL_PISTA_LLENA = 75


def horas_confirmadas(reservas, pista_id):
    return sum(1 for r in reservas if r["pista"] == pista_id and r["estado"] == "confirmada")


def porcentaje_ocupacion(horas):
    return horas * 100 / HORAS_RESERVABLES_POR_DIA


def linea_de_pista(nombre, porcentaje):
    aviso = " (llena)" if porcentaje > UMBRAL_PISTA_LLENA else ""
    return f"{nombre}: {round(porcentaje)}%{aviso}"


def informe(reservas, pistas):
    return "\n".join(
        linea_de_pista(p["nombre"], porcentaje_ocupacion(horas_confirmadas(reservas, p["id"])))
        for p in pistas
    )
