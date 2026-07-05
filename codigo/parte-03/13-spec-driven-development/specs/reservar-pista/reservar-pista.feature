# language: es
Característica: Reservar una pista
  Como socio del club
  quiero reservar una hora libre de una pista
  para asegurarme de que puedo jugar

  Antecedentes:
    Dado que el momento actual es el martes 10 de marzo a las 12:00
    Y Lucía es socia activa
    Y Inés es socia pero no está activa

  @C1
  Escenario: reserva válida
    Cuando Lucía reserva la pista Pádel 1 el miércoles 11 a las 19:00
    Entonces la reserva queda confirmada con un identificador

  @R1
  Esquema del escenario: hora fuera del horario del club
    Cuando Lucía reserva la pista Pádel 1 el miércoles 11 a las <hora>
    Entonces la reserva se rechaza con el error "fuera de horario"

    Ejemplos:
      | hora  |
      | 07:00 |
      | 22:00 |
      | 19:30 |

  @R2
  Escenario: pista ocupada
    Dado que Marcos tiene reservada la pista Pádel 1 el miércoles 11 a las 19:00
    Cuando Lucía reserva la pista Pádel 1 el miércoles 11 a las 19:00
    Entonces la reserva se rechaza con el error "pista ocupada"

  @R2
  Escenario: una reserva cancelada no ocupa la pista
    Dado que Marcos tenía reservada la pista Pádel 1 el miércoles 11 a las 19:00 y la canceló
    Cuando Lucía reserva la pista Pádel 1 el miércoles 11 a las 19:00
    Entonces la reserva queda confirmada con un identificador

  @R3
  Escenario: tercera reserva futura
    Dado que Lucía tiene reservas confirmadas el miércoles 11 y el jueves 12 a las 09:00
    Cuando Lucía reserva la pista Tenis 1 el viernes 13 a las 09:00
    Entonces la reserva se rechaza con el error "límite de reservas"

  @R4
  Escenario: socia no activa
    Cuando Inés reserva la pista Tenis 1 el miércoles 11 a las 09:00
    Entonces la reserva se rechaza con el error "socio inactivo"

  @R6
  Esquema del escenario: fuera de plazo
    Cuando Lucía reserva la pista Tenis 1 el <día> a las 09:00
    Entonces la reserva se rechaza con el error "fuera de plazo"

    Ejemplos:
      | día                 |
      | martes 10           |
      | miércoles 25        |
