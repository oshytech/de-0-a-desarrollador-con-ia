// Package dominio contiene las reglas de negocio de las reservas del club.
// No sabe nada de HTTP ni de JSON: recibe valores ya convertidos a tipos de Go.
package dominio

import (
	"errors"
	"time"
)

var (
	ErrNoEmpiezaEnPunto    = errors.New("la reserva tiene que empezar en punto")
	ErrFueraDeHorario      = errors.New("las reservas empiezan entre las 08:00 y las 21:00")
	ErrEnElPasado          = errors.New("no se puede reservar en el pasado")
	ErrDemasiadaAntelacion = errors.New("no se puede reservar con más de 14 días de antelación")
)

const (
	PrimeraHora       = 8
	UltimaHora        = 21
	DiasDeAntelacion  = 14
	DuracionDeReserva = time.Hour
)

// ValidarFranja aplica las reglas 1 y 6 del club a la hora de inicio de una
// reserva. La hora se interpreta en la zona horaria del club.
func ValidarFranja(inicio, ahora time.Time, zonaClub *time.Location) error {
	local := inicio.In(zonaClub)
	if local.Minute() != 0 || local.Second() != 0 || local.Nanosecond() != 0 {
		return ErrNoEmpiezaEnPunto
	}
	if local.Hour() < PrimeraHora || local.Hour() > UltimaHora {
		return ErrFueraDeHorario
	}
	if !inicio.After(ahora) {
		return ErrEnElPasado
	}
	if inicio.After(ahora.AddDate(0, 0, DiasDeAntelacion)) {
		return ErrDemasiadaAntelacion
	}
	return nil
}
