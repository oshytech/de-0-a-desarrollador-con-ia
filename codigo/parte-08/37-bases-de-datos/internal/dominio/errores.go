// Package dominio contiene los errores de negocio de las reservas.
// En el capítulo 47 se completa con Socio, Pista, Reserva y sus reglas.
package dominio

import "errors"

var (
	ErrSocioNoExiste   = errors.New("el socio no existe")
	ErrSocioInactivo   = errors.New("el socio no está activo")
	ErrLimiteReservas  = errors.New("el socio ya tiene 2 reservas confirmadas futuras")
	ErrPistaOcupada    = errors.New("la pista ya está reservada a esa hora")
	ErrHorarioNoValido = errors.New("la reserva no respeta el horario del club")
)
