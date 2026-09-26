// Package dominio, reducido a lo que necesita el ejemplo de revisión.
package dominio

import (
	"errors"
	"time"
)

type EstadoReserva string

const (
	Confirmada EstadoReserva = "confirmada"
	Cancelada  EstadoReserva = "cancelada"
)

type Reserva struct {
	ID      string
	SocioID string
	PistaID string
	Inicio  time.Time
	Fin     time.Time
	Estado  EstadoReserva
}

const MargenCancelacion = 2 * time.Hour

var ErrCancelacionTardia = errors.New("solo se puede cancelar hasta 2 horas antes del inicio")

// PuedeCancelar aplica la regla 5: hasta 2 horas antes del inicio.
func PuedeCancelar(r Reserva, ahora time.Time) error {
	if r.Inicio.Sub(ahora) < MargenCancelacion {
		return ErrCancelacionTardia
	}
	return nil
}
