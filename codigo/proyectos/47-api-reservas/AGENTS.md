# Instrucciones para agentes: API de reservas del club

## Qué es
API REST en Go para reservar pistas de un club. Contrato: `openapi.yaml`.
Reglas del club: las seis de la descripción de `openapi.yaml`.

## Estructura (no la cambies sin preguntar)
- `internal/dominio`: entidades y reglas. Solo librería estándar.
- `internal/aplicacion`: casos de uso y puertos. Solo depende del dominio.
- `internal/adaptadores/{memoria,postgres,http,reloj}`: implementaciones.
- `cmd/api/main.go`: composición. Sin reglas de negocio.

## Cómo se trabaja
- Un paso cada vez. Cuando termines lo que se te pide, para y resume qué has cambiado.
- Los tests existentes son la especificación: no los cambies para que pase el código.
  Si crees que un test está mal, dilo y espera.
- Nada de dependencias nuevas salvo `github.com/jackc/pgx/v5`.
- Errores de dominio: los define `dominio` y los traducen los adaptadores.

## Comandos
- Tests: `go test ./...`
- Con PostgreSQL: `docker compose up -d --wait` y
  `RESERVAS_BD=postgres://reservas:reservas_local@localhost:5432/reservas go test ./...`
- No ejecutes nada contra otra base de datos que la del `docker-compose.yml`.
