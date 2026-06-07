"""Resumen semanal de pasos: el ejemplo que recorre el capítulo 6."""

OBJETIVO_DIARIO = 10000

pasos_por_dia = {
    "lunes": 8250,
    "martes": 11020,
    "miércoles": 6400,
    "jueves": 12500,
    "viernes": 9800,
    "sábado": 15030,
    "domingo": 4300,
}


def media(valores):
    """Devuelve la media de una lista de números."""
    if not valores:
        raise ValueError("No se puede calcular la media de una lista vacía")
    return sum(valores) / len(valores)


def dias_que_cumplen(pasos, objetivo):
    """Devuelve la lista de días en los que se llegó al objetivo."""
    cumplidos = []
    for dia, cantidad in pasos.items():
        if cantidad >= objetivo:
            cumplidos.append(dia)
    return cumplidos


def mejor_dia(pasos):
    """Devuelve el día con más pasos y cuántos fueron."""
    dia_maximo = None
    maximo = 0
    for dia, cantidad in pasos.items():
        if cantidad > maximo:
            dia_maximo = dia
            maximo = cantidad
    return dia_maximo, maximo


def valorar(media_diaria):
    """Traduce la media diaria a un comentario."""
    if media_diaria >= 12000:
        return "semana muy activa"
    elif media_diaria >= OBJETIVO_DIARIO:
        return "objetivo cumplido de media"
    elif media_diaria >= 7000:
        return "cerca del objetivo"
    else:
        return "semana tranquila"


media_semana = media(list(pasos_por_dia.values()))
cumplidos = dias_que_cumplen(pasos_por_dia, OBJETIVO_DIARIO)
dia, cantidad = mejor_dia(pasos_por_dia)

print(f"Media diaria: {media_semana:.0f} pasos ({valorar(media_semana)})")
print(f"Días con objetivo: {len(cumplidos)} de {len(pasos_por_dia)}: {', '.join(cumplidos)}")
print(f"Mejor día: {dia}, con {cantidad} pasos")
