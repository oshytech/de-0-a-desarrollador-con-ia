# Reservas del club

API y herramientas para un club de barrio que alquila pistas de tenis y pádel.
Las reglas de negocio están en docs/reglas.md y son la fuente de verdad.

## Comandos
- Tests: `go test ./...` desde la raíz. Tienen que pasar antes de dar nada por terminado.
- Servidor MCP de pruebas: `python -I probar_servidor.py`.

## Mapa
- `internal/dominio/`: entidades y reglas. Sin dependencias externas.
- `internal/aplicacion/`: casos de uso y puertos.
- `internal/adaptadores/`: HTTP, PostgreSQL y memoria. Cada carpeta tiene su propio AGENTS.md.
- `datos/`: datos ficticios. Solo lectura.

## Cuándo leer qué
- Si el cambio toca una regla de reserva, lee docs/reglas.md antes de proponer nada.
- Si te piden un informe de ocupación, usa la skill informe-ocupacion.
- Para cualquier texto que lean los socios, lee docs/tono.md.

## Siempre
- Identificadores y mensajes en español, como el resto del código.
- Un test nuevo por cada regla que cambie, con el caso límite (las 21:00, las 2 horas justas).

## Pregunta antes
- De añadir una dependencia.
- De cambiar el esquema de la base de datos o una migración ya aplicada.

## Nunca
- Borrar o desactivar un test que falla para que el resto pase.
- Leer ni escribir ficheros `.env`. Si necesitas una variable, pregunta su nombre.
- Hacer push ni tocar las ramas compartidas.

## Al terminar
Explica qué ficheros has cambiado, qué comandos has ejecutado y su resultado,
y qué no has podido comprobar.
