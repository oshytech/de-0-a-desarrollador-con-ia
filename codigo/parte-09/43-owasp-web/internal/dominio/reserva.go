// Package dominio: reservas del club y la regla de cancelación.
package dominio

import (
	"errors"
	"time"
)

type Estado string

const (
	Confirmada Estado = "confirmada"
	Cancelada  Estado = "cancelada"
)

type Reserva struct {
	ID      int64
	SocioID int64
	PistaID string
	Inicio  time.Time
	Estado  Estado
}

var (
	ErrReservaNoEncontrada = errors.New("reserva no encontrada")
	ErrFueraDePlazo        = errors.New("solo se puede cancelar hasta 2 horas antes del inicio")
	ErrYaCancelada         = errors.New("la reserva ya estaba cancelada")
)

const PlazoDeCancelacion = 2 * time.Hour

// Cancelar aplica la regla 5: se puede cancelar hasta 2 horas antes.
func (r *Reserva) Cancelar(ahora time.Time) error {
	if r.Estado == Cancelada {
		return ErrYaCancelada
	}
	if ahora.After(r.Inicio.Add(-PlazoDeCancelacion)) {
		return ErrFueraDePlazo
	}
	r.Estado = Cancelada
	return nil
}
