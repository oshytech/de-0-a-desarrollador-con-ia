# Spec: reservar una pista

Estado: aprobada · Sale de HU2 de `requisitos-club.md`

## Objetivo

Un socio activo reserva una hora libre de una pista y el sistema aplica todas las reglas del club sin intervención de la encargada.

## Fuera de alcance

- Cancelar reservas (tiene su propia spec).
- Avisos por correo: el caso de uso publica que la reserva se ha creado y otra pieza decide cómo avisar.
- Interfaz web. Esta spec cubre el caso de uso y sus errores.

## Glosario

- **Reserva confirmada**: reserva en estado `confirmada`. Las canceladas no cuentan para ninguna regla.
- **Reserva futura**: reserva cuyo inicio es posterior al momento actual del sistema.
- **Momento actual**: el que da el reloj del sistema. En los tests, un reloj fijo.

## Reglas y criterios de aceptación

- **R1.** CUANDO un socio pida una reserva que no empiece en punto, o que empiece antes de las 08:00 o después de las 21:00, EL sistema DEBERÁ rechazarla con el error `fuera de horario`.
- **R2.** CUANDO la pista tenga una reserva confirmada que se solape con la hora pedida, EL sistema DEBERÁ rechazarla con el error `pista ocupada`.
- **R3.** CUANDO el socio tenga ya 2 reservas confirmadas futuras, EL sistema DEBERÁ rechazarla con el error `límite de reservas`.
- **R4.** CUANDO el socio no esté activo, EL sistema DEBERÁ rechazarla con el error `socio inactivo`.
- **R6.** CUANDO la hora pedida sea anterior o igual al momento actual, o posterior en más de 14 días, EL sistema DEBERÁ rechazarla con el error `fuera de plazo`.
- **C1.** CUANDO se cumplan todas las reglas, EL sistema DEBERÁ guardar la reserva como confirmada, asignarle un identificador y devolverla.

R5 (cancelación) no aplica a esta spec.

## Orden de comprobación

Si una petición incumple varias reglas, se devuelve el primer error en este orden: R4, R1, R6, R2, R3. Así el socio inactivo siempre recibe el mismo mensaje, haga lo que haga.

## Preguntas resueltas

- ¿Cuenta como solapada una reserva cancelada? No (glosario).
- ¿Se puede reservar a las 21:30? No: las reservas empiezan en punto (R1).
