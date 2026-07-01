# Reservas de pistas del club: requisitos

Versión 0.3 · Documento vivo: cada cambio queda en el historial de Git.

## Contexto

El club tiene cuatro pistas (dos de tenis y dos de pádel) y unos trescientos socios. Hoy las reservas se hacen por teléfono y por un grupo de mensajería, y la encargada las apunta en una hoja de cálculo. Hay pistas reservadas dos veces, socios que acaparan las mejores horas y reservas que nadie usa porque se cancelan por mensaje y no llegan a la hoja.

## Objetivos

1. Que un socio pueda reservar una pista sin llamar a nadie.
2. Que no haya dos reservas para la misma pista a la misma hora.
3. Que las horas buenas se repartan entre más socios.

## No objetivos (de momento)

- Cobrar las reservas. Las pistas están incluidas en la cuota.
- Torneos, clases y escuela de pádel.
- Aplicación móvil nativa: basta con una web que funcione bien en el móvil.

## Antiobjetivos (si pasa esto, ha fallado aunque funcione)

- Que la encargada tenga que seguir apuntando reservas a mano.
- Que reservar cueste más pasos que mandar un mensaje.

## Actores

- **Socio**: consulta disponibilidad, reserva y cancela sus reservas.
- **Encargada**: da de alta y de baja a los socios y consulta todas las reservas.
- **Sistema**: aplica las reglas, guarda las reservas y envía los avisos.

## Reglas de negocio

| ID | Regla | Fuente |
|---|---|---|
| R1 | Las reservas duran una hora exacta y empiezan en punto, entre las 08:00 y las 22:00 (la última empieza a las 21:00). | Horario del club |
| R2 | Una pista no puede tener dos reservas confirmadas que se solapen. | Entrevista con la encargada |
| R3 | Un socio no puede tener más de 2 reservas confirmadas futuras. | Acta de la junta |
| R4 | Solo los socios activos pueden reservar. | Estatutos |
| R5 | Se puede cancelar hasta 2 horas antes del inicio. Después, no. | Acta de la junta |
| R6 | No se reserva en el pasado ni con más de 14 días de antelación. | Acta de la junta |

## Historias de usuario

| ID | Historia | Prioridad |
|---|---|---|
| HU1 | Como socio, quiero ver qué horas están libres en cada pista para elegir cuándo jugar. | Must |
| HU2 | Como socio, quiero reservar una hora libre para asegurarme la pista. | Must |
| HU3 | Como socio, quiero cancelar una reserva para que otro pueda usar la pista. | Must |
| HU4 | Como encargada, quiero dar de baja a un socio para que deje de poder reservar. | Must |
| HU5 | Como socio, quiero recibir un aviso cuando reservo o cancelo para tener constancia. | Should |
| HU6 | Como socio, quiero apuntarme a una lista de espera de una hora ocupada para jugar si alguien cancela. | Could |
| HU7 | Como socio, quiero reservar la misma hora todas las semanas sin repetir el proceso. | Won't (por ahora) |

### Criterios de aceptación de HU2

```gherkin
Escenario: reserva de una hora libre
  Dado que Lucía es socia activa y no tiene reservas
  Y la pista Pádel 1 está libre el martes a las 19:00
  Cuando Lucía reserva la pista Pádel 1 el martes a las 19:00
  Entonces la reserva queda confirmada
  Y la hora deja de aparecer como libre para los demás socios

Escenario: hora ya ocupada
  Dado que Marcos tiene reservada la pista Pádel 1 el martes a las 19:00
  Cuando Lucía intenta reservar la pista Pádel 1 el martes a las 19:00
  Entonces la reserva se rechaza con el motivo "la pista ya está reservada a esa hora"
```

## Requisitos no funcionales

| ID | Requisito | Cómo se comprueba |
|---|---|---|
| RNF1 | La lista de horas libres de un día se muestra en menos de 1 segundo en el percentil 95. | Medición en el entorno de pruebas con 300 socios y 14 días de reservas |
| RNF2 | Reservar no exige más de 3 pulsaciones desde la pantalla de inicio. | Prueba con cinco socios reales |
| RNF3 | Los correos y teléfonos de los socios solo los ve la encargada. | Test de permisos por rol |

## Restricciones

- Presupuesto de alojamiento: el del hosting que ya paga el club.
- Tiene que estar en marcha antes del inicio de la temporada de verano.

## Preguntas abiertas

- ¿Una reserva puede ser de dos socios (partido de dobles) o solo cuenta para quien la hace? *Decisión provisional: cuenta solo para quien reserva.*
- ¿Qué pasa con las reservas de un socio al que se da de baja? *Pendiente de la junta.*
