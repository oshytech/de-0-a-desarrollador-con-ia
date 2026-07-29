# language: es
Característica: Reservar una pista
  Como socio del club
  quiero reservar una hora de pista
  para jugar sin tener que llamar a recepción

  Antecedentes:
    Dado que son las 09:30 del 1 de diciembre de 2026
    Y que Ana es socia activa
    Y que la pista 1 es de tenis

  Escenario: La pista está libre
    Cuando Ana reserva la pista 1 el 2 de diciembre a las 10:00
    Entonces la reserva queda confirmada de 10:00 a 11:00

  Escenario: La pista ya está reservada a esa hora
    Dado que Marta tiene la pista 1 el 2 de diciembre a las 10:00
    Cuando Ana reserva la pista 1 el 2 de diciembre a las 10:00
    Entonces la reserva se rechaza porque la pista está ocupada

  Escenario: La hora no empieza en punto
    Cuando Ana reserva la pista 1 el 2 de diciembre a las 10:30
    Entonces la reserva se rechaza porque está fuera de horario

  Escenario: Demasiada antelación
    Cuando Ana reserva la pista 1 el 2 de enero de 2027 a las 10:00
    Entonces la reserva se rechaza porque está fuera de plazo

  Escenario: Ya tiene dos reservas futuras
    Dado que Ana tiene reservas confirmadas el 2 y el 3 de diciembre
    Cuando Ana reserva la pista 2 el 4 de diciembre a las 10:00
    Entonces la reserva se rechaza porque ha llegado al límite de reservas

  Escenario: Socio de baja
    Dado que Luis es socio pero no está activo
    Cuando Luis reserva la pista 1 el 2 de diciembre a las 10:00
    Entonces la reserva se rechaza porque el socio no está activo
