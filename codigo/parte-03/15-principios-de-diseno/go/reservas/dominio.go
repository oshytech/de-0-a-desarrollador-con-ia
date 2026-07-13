// Package reservas reúne, en un solo paquete, el dominio y el caso de uso de reservar.
// En la parte IV se reparte en capas.
package reservas

import (
	"errors"
	"time"
)

const (
	Apertura           = 8
	UltimoInicio       = 21
	Duracion           = time.Hour
	AntelacionMaxima   = 14 * 24 * time.Hour
	MaxReservasFuturas = 2
)

type Socio struct {
	ID     int
	Nombre string
	Correo string
	Activo bool
}

type Reserva struct {
	ID      int
	SocioID int
	PistaID int
	Inicio  time.Time
	Estado  string
}

func (r Reserva) Fin() time.Time { return r.Inicio.Add(Duracion) }

func (r Reserva) SeSolapaCon(inicio time.Time) bool {
	return r.Estado == "confirmada" && r.Inicio.Before(inicio.Add(Duracion)) && inicio.Before(r.Fin())
}

var (
	ErrSocioNoExiste  = errors.New("el socio no existe")
	ErrSocioInactivo  = errors.New("el socio no está activo")
	ErrFueraDeHorario = errors.New("las reservas empiezan en punto, entre las 08:00 y las 21:00")
	ErrFueraDePlazo   = errors.New("no se reserva en el pasado ni con más de 14 días de antelación")
	ErrPistaOcupada   = errors.New("la pista ya está reservada a esa hora")
	ErrLimiteReservas = errors.New("el socio ya tiene 2 reservas futuras")
)
