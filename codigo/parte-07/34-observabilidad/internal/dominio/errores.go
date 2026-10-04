// Package dominio contiene los errores de negocio de las reservas del club.
// En este ejemplo solo hacen falta los errores: el capítulo trata de cómo se
// observan, no de cómo se comprueban las reglas.
package dominio

import "errors"

// Errores de dominio. Son rechazos esperados, no fallos del sistema.
var (
	ErrHorarioNoValido   = errors.New("la reserva debe empezar en punto entre las 08:00 y las 21:00")
	ErrPistaOcupada      = errors.New("la pista ya tiene una reserva confirmada a esa hora")
	ErrLimiteReservas    = errors.New("el socio ya tiene 2 reservas confirmadas futuras")
	ErrSocioInactivo     = errors.New("el socio no está activo")
	ErrFueraDePlazo      = errors.New("no se reserva en el pasado ni con más de 14 días de antelación")
	ErrCancelacionTardia = errors.New("solo se puede cancelar hasta 2 horas antes del inicio")
)
