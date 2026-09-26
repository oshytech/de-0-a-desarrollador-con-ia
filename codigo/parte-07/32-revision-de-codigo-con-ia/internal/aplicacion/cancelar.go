// Package aplicacion contiene los casos de uso y los puertos que necesitan.
package aplicacion

import (
	"errors"
	"time"

	"reservas/internal/dominio"
)

var ErrReservaAjena = errors.New("la reserva no es de este socio")

type RepositorioReservas interface {
	Buscar(id string) (dominio.Reserva, error)
	Guardar(r dominio.Reserva) error
}

type Reloj interface {
	Ahora() time.Time
}

type CancelarReserva struct {
	Reservas RepositorioReservas
	Reloj    Reloj
}

// Ejecutar cancela la reserva si es del socio y aún está en plazo.
// Cancelar una reserva ya cancelada no hace nada y no es un error.
func (c CancelarReserva) Ejecutar(socioID, reservaID string) error {
	r, err := c.Reservas.Buscar(reservaID)
	if err != nil {
		return err
	}
	if r.SocioID != socioID {
		return ErrReservaAjena
	}
	if r.Estado == dominio.Cancelada {
		return nil
	}
	if err := dominio.PuedeCancelar(r, c.Reloj.Ahora()); err != nil {
		return err
	}
	r.Estado = dominio.Cancelada
	return c.Reservas.Guardar(r)
}
