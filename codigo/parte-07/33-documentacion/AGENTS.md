# Instrucciones para agentes

- Lenguaje: Go 1.27, solo librería estándar. Identificadores en español.
- Antes de dar un cambio por terminado: `go vet ./...` y `go test ./...` en verde.
- Toda función exportada lleva comentario de documentación que empieza por su nombre.
- Si cambias una regla de negocio, actualiza en el mismo cambio el comentario del paquete `dominio`, el ejemplo afectado en `example_test.go` y, si cambia el flujo, `docs/flujo-reserva.md`.
- `internal/dominio` no importa nada de fuera de la librería estándar ni sabe de HTTP o SQL (ver `docs/adr/0001-...`).
- Las decisiones de arquitectura nuevas se proponen como ADR en `docs/adr/`. No borres ADR antiguos: márcalos como sustituidos.
