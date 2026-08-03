package dominio

import "time"

// Evento es un hecho del negocio que ya ha ocurrido. Se nombra en pasado.
type Evento interface {
	Nombre() string
	// Clave identifica el hecho. Si llega dos veces, es el mismo hecho
	// y quien lo consume puede descartar la copia.
	Clave() string
}

type ReservaConfirmada struct {
	ReservaID string
	SocioID   string
	PistaID   string
	Inicio    time.Time
}

func (e ReservaConfirmada) Nombre() string { return "reserva.confirmada" }
func (e ReservaConfirmada) Clave() string  { return e.Nombre() + ":" + e.ReservaID }

type ReservaCancelada struct {
	ReservaID string
	SocioID   string
	PistaID   string
	Inicio    time.Time
}

func (e ReservaCancelada) Nombre() string { return "reserva.cancelada" }
func (e ReservaCancelada) Clave() string  { return e.Nombre() + ":" + e.ReservaID }

// Una reserva se confirma una vez y se cancela como mucho una vez,
// así que el par (nombre, reserva) basta para reconocer un duplicado.

func ConfirmacionDe(r Reserva) ReservaConfirmada {
	return ReservaConfirmada{ReservaID: r.id, SocioID: r.socioID, PistaID: r.pistaID, Inicio: r.franja.Inicio()}
}

func CancelacionDe(r Reserva) ReservaCancelada {
	return ReservaCancelada{ReservaID: r.id, SocioID: r.socioID, PistaID: r.pistaID, Inicio: r.franja.Inicio()}
}
