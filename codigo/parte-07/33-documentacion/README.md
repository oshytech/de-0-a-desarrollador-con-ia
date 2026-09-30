# Reservas del club

Dominio de reservas de pistas de tenis y pádel de un club de barrio. Este directorio es el ejemplo del capítulo 33: muestra cómo se documenta un paquete de Go para que la documentación se compruebe sola.

## Requisitos

- Go 1.27 o posterior. No hay dependencias externas.

## Probarlo

```bash
go test ./...                        # tests y ejemplos ejecutables
go doc ./internal/dominio            # documentación del paquete
go doc ./internal/dominio NuevaReserva
```

Si un ejemplo deja de coincidir con lo que hace el código, `go test` falla.

## Estructura

```
internal/dominio/       entidades, errores y reglas que no necesitan base de datos
docs/flujo-reserva.md   diagrama del flujo de una reserva (Mermaid)
docs/adr/               decisiones de arquitectura
AGENTS.md               instrucciones para agentes de IA que trabajen aquí
```

## Reglas de negocio

Están en el comentario del paquete `dominio` (`go doc ./internal/dominio`). No se copian aquí para que haya un único sitio que mantener.
