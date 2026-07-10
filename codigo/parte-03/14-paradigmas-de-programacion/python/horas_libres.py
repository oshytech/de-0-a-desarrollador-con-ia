"""Horas libres de una pista: la misma pregunta resuelta de forma imperativa y declarativa."""

APERTURA = 8        # primera hora en la que puede empezar una reserva
ULTIMO_INICIO = 21  # la última reserva empieza a las 21:00 y acaba a las 22:00


def horas_libres_imperativo(ocupadas):
    libres = []
    hora = APERTURA
    while hora <= ULTIMO_INICIO:
        if hora not in ocupadas:
            libres.append(hora)
        hora += 1
    return libres


def horas_libres_declarativo(ocupadas):
    return [h for h in range(APERTURA, ULTIMO_INICIO + 1) if h not in ocupadas]


if __name__ == "__main__":
    ocupadas = {9, 10, 18, 19, 20}
    print(horas_libres_imperativo(ocupadas))
    print(horas_libres_declarativo(ocupadas))
