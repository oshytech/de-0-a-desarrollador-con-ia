package dominio

import "time"

const PlazoCancelacion = 2 * time.Hour

// Cancelar aplica la regla 5 y devuelve la reserva ya cancelada.
// No modifica la original: quien llama decide si guarda el resultado.
func (r Reserva) Cancelar(socioID string, ahora time.Time) (Reserva, error) {
	if r.estado == Cancelada {
		return Reserva{}, ErrReservaYaCancelada
	}
	if r.socioID != socioID {
		return Reserva{}, ErrReservaDeOtroSocio
	}
	if r.franja.Inicio().Sub(ahora) < PlazoCancelacion {
		return Reserva{}, ErrCancelacionFueraDePlazo
	}
	r.estado = Cancelada
	return r, nil
}
