# Observabilidad del servidor de reservas

Ejemplo del capítulo 34. Solo librería estándar de Go.

```bash
go test ./...          # middleware, salud, percentiles y alertas
go run ./cmd/demo      # logs JSON reales, métricas de 200 peticiones y alertas
```

- `internal/adaptadores/http/observar.go`: middleware con identificador de petición, logs con `log/slog` y registro de duraciones.
- `internal/adaptadores/http/manejadores.go`: `POST /reservas` (rechazos de negocio frente a errores del sistema) y `GET /health`.
- `internal/metricas`: contadores y percentiles en memoria (didáctico: en producción, histogramas).
- `internal/alertas`: alerta por duración frente a una referencia y agrupación de alertas repetidas.
