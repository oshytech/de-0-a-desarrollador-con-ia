// Package aplicacion contiene los casos de uso y los puertos que necesitan.
package aplicacion

import (
	"context"
	"time"

	"owaspweb/internal/dominio"
)

type RepositorioReservas interface {
	Buscar(ctx context.Context, id int64) (dominio.Reserva, error)
	Guardar(ctx context.Context, r dominio.Reserva) error
}

type Reloj interface{ Ahora() time.Time }

// CancelarReserva comprueba las tres preguntas del control de acceso:
// quién lo pide (socioID, que llega de la sesión), qué quiere hacer
// (cancelar) y sobre qué recurso (la reserva y su dueño).
type CancelarReserva struct {
	Reservas RepositorioReservas
	Reloj    Reloj
}

func (c CancelarReserva) Ejecutar(ctx context.Context, socioID, reservaID int64) error {
	r, err := c.Reservas.Buscar(ctx, reservaID)
	if err != nil {
		return err
	}
	if r.SocioID != socioID {
		// Misma respuesta que si no existiera: no confirmamos a nadie que
		// la reserva 1043 existe y es de otro socio.
		return dominio.ErrReservaNoEncontrada
	}
	if err := r.Cancelar(c.Reloj.Ahora()); err != nil {
		return err
	}
	return c.Reservas.Guardar(ctx, r)
}
