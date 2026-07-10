"""Python comprueba los tipos al ejecutar: el error aparece solo cuando se llega a esa línea."""


def fin_de_reserva(hora_inicio):
    return hora_inicio + 1


print(fin_de_reserva(9))
print(fin_de_reserva("9"))
