// Package dominio contiene las entidades y las reglas del club, sin dependencias externas.
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

var (
	ErrHorarioNoValido   = errors.New("horario no válido")
	ErrPistaOcupada      = errors.New("la pista ya está reservada en esa hora")
	ErrLimiteReservas    = errors.New("el socio ya tiene el máximo de reservas futuras")
	ErrSocioInactivo     = errors.New("el socio no está activo")
	ErrFueraDePlazo      = errors.New("fuera del plazo de reserva")
	ErrCancelacionTardia = errors.New("ya no se puede cancelar")
)
