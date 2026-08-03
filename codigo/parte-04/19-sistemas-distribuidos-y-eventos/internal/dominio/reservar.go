package dominio

import "time"

const (
	MaxReservasFuturas = 2
	MaxAntelacion      = 14 * 24 * time.Hour
)

// Ocupacion es lo que el caso de uso ha averiguado antes de pedir la reserva.
// El dominio no sabe de dónde sale: solo decide con ello.
type Ocupacion struct {
	DeLaPista       []Reserva // reservas de la pista que podrían solaparse
	FuturasDelSocio []Reserva // reservas del socio a partir de ahora
}

// NuevaReserva aplica las reglas de creación en el orden de la spec
// (R4, R1, R6, R2, R3) y devuelve una reserva confirmada.
func NuevaReserva(id string, socio Socio, pista Pista, inicio, ahora time.Time, ocupacion Ocupacion) (Reserva, error) {
	if err := socio.PuedeReservar(); err != nil { // R4
		return Reserva{}, err
	}
	franja, err := NuevaFranja(inicio) // R1
	if err != nil {
		return Reserva{}, err
	}
	if !inicio.After(ahora) || inicio.Sub(ahora) > MaxAntelacion { // R6
		return Reserva{}, ErrFueraDePlazo
	}
	for _, r := range ocupacion.DeLaPista { // R2
		if r.pistaID == pista.ID && r.EstaConfirmada() && r.franja.SeSolapaCon(franja) {
			return Reserva{}, ErrPistaOcupada
		}
	}
	futuras := 0
	for _, r := range ocupacion.FuturasDelSocio { // R3
		if r.socioID == socio.ID && r.EstaConfirmada() && r.franja.Inicio().After(ahora) {
			futuras++
		}
	}
	if futuras >= MaxReservasFuturas {
		return Reserva{}, ErrLimiteDeReservas
	}
	return Reserva{id: id, socioID: socio.ID, pistaID: pista.ID, franja: franja, estado: Confirmada}, nil
}
