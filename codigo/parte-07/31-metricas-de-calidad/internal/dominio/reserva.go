// Package dominio contiene las entidades y reglas del club de pistas.
// No depende de nada fuera de la librería estándar.
package dominio

import (
	"errors"
	"time"
)

type Socio struct {
	ID     string
	Nombre string
	Correo string
	Activo bool
}

type Pista struct {
	ID      string
	Nombre  string
	Deporte string
}

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

const (
	Duracion           = time.Hour
	PrimeraHora        = 8
	UltimaHora         = 21 // la última reserva empieza a las 21:00
	MaxReservasFuturas = 2
	AntelacionMaxima   = 14 * 24 * time.Hour
	MargenCancelacion  = 2 * time.Hour
)

var (
	ErrHorarioNoValido   = errors.New("la reserva tiene que empezar en punto entre las 08:00 y las 21:00")
	ErrPistaOcupada      = errors.New("la pista ya está reservada a esa hora")
	ErrLimiteReservas    = errors.New("el socio ya tiene el máximo de reservas futuras")
	ErrSocioInactivo     = errors.New("solo los socios activos pueden reservar")
	ErrFueraDePlazo      = errors.New("no se reserva en el pasado ni con más de 14 días de antelación")
	ErrCancelacionTardia = errors.New("solo se puede cancelar hasta 2 horas antes del inicio")
)

// ValidarHorario comprueba las reglas 1 y 6: hora en punto, franja del club y plazo.
func ValidarHorario(inicio, ahora time.Time) error {
	if inicio.Minute() != 0 || inicio.Second() != 0 || inicio.Nanosecond() != 0 {
		return ErrHorarioNoValido
	}
	if inicio.Hour() < PrimeraHora || inicio.Hour() > UltimaHora {
		return ErrHorarioNoValido
	}
	if !inicio.After(ahora) || inicio.Sub(ahora) > AntelacionMaxima {
		return ErrFueraDePlazo
	}
	return nil
}

// PuedeReservar aplica todas las reglas de una reserva nueva.
// dePista son las reservas de la pista y deSocio, las del socio.
func PuedeReservar(socio Socio, inicio, ahora time.Time, dePista, deSocio []Reserva) error {
	if !socio.Activo {
		return ErrSocioInactivo
	}
	if err := ValidarHorario(inicio, ahora); err != nil {
		return err
	}
	if pistaOcupada(dePista, inicio) {
		return ErrPistaOcupada
	}
	if reservasFuturas(deSocio, ahora) >= MaxReservasFuturas {
		return ErrLimiteReservas
	}
	return nil
}

// pistaOcupada dice si alguna reserva confirmada se solapa con la hora que empieza en inicio.
func pistaOcupada(dePista []Reserva, inicio time.Time) bool {
	fin := inicio.Add(Duracion)
	for _, r := range dePista {
		if r.Estado == Confirmada && r.Inicio.Before(fin) && inicio.Before(r.Fin) {
			return true
		}
	}
	return false
}

// reservasFuturas cuenta las reservas confirmadas que todavía no han empezado.
func reservasFuturas(deSocio []Reserva, ahora time.Time) int {
	n := 0
	for _, r := range deSocio {
		if r.Estado == Confirmada && r.Inicio.After(ahora) {
			n++
		}
	}
	return n
}

// PuedeCancelar aplica la regla 5: hasta 2 horas antes del inicio.
func PuedeCancelar(r Reserva, ahora time.Time) error {
	if r.Inicio.Sub(ahora) < MargenCancelacion {
		return ErrCancelacionTardia
	}
	return nil
}
