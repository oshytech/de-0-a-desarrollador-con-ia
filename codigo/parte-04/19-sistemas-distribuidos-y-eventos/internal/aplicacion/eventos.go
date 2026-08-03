package aplicacion

import (
	"context"

	"reservas/internal/dominio"
)

// PublicadorEventos es el puerto por el que los casos de uso cuentan lo que
// ha pasado. No dice quién escucha ni cómo llega: eso lo decide el adaptador.
type PublicadorEventos interface {
	Publicar(ctx context.Context, eventos ...dominio.Evento) error
}

// Notificador envía un mensaje a un socio. Hoy escribe en la consola;
// mañana puede ser un servicio de correo.
type Notificador interface {
	Enviar(ctx context.Context, correo, asunto, texto string) error
}
