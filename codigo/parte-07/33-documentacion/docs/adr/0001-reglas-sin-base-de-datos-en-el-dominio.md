# ADR 0001. Las reglas que no necesitan datos viven en el dominio

- Estado: aceptada
- Fecha: 2026-03-10

## Contexto

Las reglas de reserva se pueden comprobar en tres sitios: en el handler HTTP, en el caso de uso o en las propias entidades. Algunas solo necesitan la reserva y la hora actual (horario, plazo, cancelación). Otras necesitan consultar reservas existentes (solapes, límite por socio).

## Opciones

1. Todas las reglas en el caso de uso.
2. Las reglas sin datos en el dominio y las que consultan reservas en el caso de uso.
3. Validación en el handler HTTP.

## Decisión

Opción 2. `NuevaReserva` y `Cancelar` comprueban lo que no necesita base de datos. `ReservarPista` comprueba solapes y límite usando los repositorios.

## Consecuencias

- Las reglas del dominio se prueban sin repositorios ni dobles de prueba.
- Las reglas quedan repartidas en dos paquetes. El comentario del paquete `dominio` dice dónde está cada una.
- Validar en el handler queda descartado: otra entrada (una tarea programada, una CLI) se saltaría las reglas.
