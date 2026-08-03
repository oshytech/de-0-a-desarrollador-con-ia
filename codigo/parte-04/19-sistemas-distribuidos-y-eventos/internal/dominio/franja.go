package dominio

import "time"

const (
	PrimeraHora = 8  // la primera reserva empieza a las 08:00
	UltimaHora  = 21 // la última empieza a las 21:00 y acaba a las 22:00
	Duracion    = time.Hour
)

// Franja es la hora de pista que ocupa una reserva. Es un value object:
// no tiene identidad, se define por su valor y, si existe, es válida.
type Franja struct {
	inicio time.Time
}

// NuevaFranja aplica la regla 1: una hora exacta que empieza en punto.
func NuevaFranja(inicio time.Time) (Franja, error) {
	enPunto := inicio.Minute() == 0 && inicio.Second() == 0 && inicio.Nanosecond() == 0
	if !enPunto || inicio.Hour() < PrimeraHora || inicio.Hour() > UltimaHora {
		return Franja{}, ErrFueraDeHorario
	}
	return Franja{inicio: inicio}, nil
}

func (f Franja) Inicio() time.Time { return f.inicio }
func (f Franja) Fin() time.Time    { return f.inicio.Add(Duracion) }

// SeSolapaCon dice si dos franjas comparten algún instante.
func (f Franja) SeSolapaCon(otra Franja) bool {
	return f.Inicio().Before(otra.Fin()) && otra.Inicio().Before(f.Fin())
}
