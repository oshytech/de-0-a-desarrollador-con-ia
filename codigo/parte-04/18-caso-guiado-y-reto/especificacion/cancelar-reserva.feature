# language: es
Característica: Cancelar una reserva (el reto)
  Como socio del club
  quiero cancelar una reserva que ya no voy a usar
  para que otro socio pueda aprovechar la pista

  Antecedentes:
    Dado que Ana tiene la pista 1 el 2 de diciembre de 2026 a las 10:00

  Escenario: Cancela con tiempo
    Dado que son las 07:59 del 2 de diciembre
    Cuando Ana cancela su reserva
    Entonces la reserva queda cancelada
    Y la pista 1 vuelve a estar libre a las 10:00

  Escenario: Justo en el límite
    Dado que son las 08:00 del 2 de diciembre
    Cuando Ana cancela su reserva
    Entonces la reserva queda cancelada

  Escenario: Demasiado tarde
    Dado que son las 08:01 del 2 de diciembre
    Cuando Ana cancela su reserva
    Entonces la cancelación se rechaza porque quedan menos de 2 horas

  Escenario: La reserva no es suya
    Cuando Marta intenta cancelar la reserva de Ana
    Entonces la cancelación se rechaza porque la reserva es de otro socio

  Escenario: Ya estaba cancelada
    Dado que Ana ya canceló su reserva
    Cuando Ana vuelve a cancelarla
    Entonces la cancelación se rechaza porque ya estaba cancelada
