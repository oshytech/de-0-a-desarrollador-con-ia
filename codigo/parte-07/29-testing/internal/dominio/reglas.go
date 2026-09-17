package dominio

import "time"

const (
	HoraApertura                = 8
	UltimaHoraInicio            = 21
	DuracionReserva             = time.Hour
	AntelacionMaxima            = 14 * 24 * time.Hour
	AntelacionMinimaCancelacion = 2 * time.Hour
	MaxReservasFuturas          = 2
)

// ComprobarHorario aplica las reglas 1 y 6: en punto, entre las 08:00 y las 21:00,
// ni en el pasado ni con más de 14 días de antelación.
func ComprobarHorario(inicio, ahora time.Time) error {
	if inicio.Minute() != 0 || inicio.Second() != 0 || inicio.Nanosecond() != 0 {
		return ErrHorarioNoValido
	}
	hora := inicio.Hour()
	if hora < HoraApertura || hora > UltimaHoraInicio {
		return ErrHorarioNoValido
	}
	if !inicio.After(ahora) || inicio.Sub(ahora) > AntelacionMaxima {
		return ErrFueraDePlazo
	}
	return nil
}

// SeSolapan aplica la regla 2 a dos reservas de la misma pista.
func SeSolapan(a, b Reserva) bool {
	return a.Inicio.Before(b.Fin) && b.Inicio.Before(a.Fin)
}

// ComprobarLimite aplica la regla 3.
func ComprobarLimite(confirmadasFuturas int) error {
	if confirmadasFuturas >= MaxReservasFuturas {
		return ErrLimiteReservas
	}
	return nil
}

// ComprobarCancelacion aplica la regla 5: se puede cancelar hasta 2 horas antes.
func ComprobarCancelacion(r Reserva, ahora time.Time) error {
	if r.Inicio.Sub(ahora) < AntelacionMinimaCancelacion {
		return ErrCancelacionTardia
	}
	return nil
}
