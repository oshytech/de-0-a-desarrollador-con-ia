package dominio

import "time"

type Estado string

const (
	Confirmada Estado = "confirmada"
	Cancelada  Estado = "cancelada"
)

// Reserva tiene un ciclo de vida, así que sus campos no se exportan:
// solo cambia a través de sus métodos.
type Reserva struct {
	id      string
	socioID string
	pistaID string
	franja  Franja
	estado  Estado
}

// ReconstruirReserva vuelve a montar una reserva que ya existía, por ejemplo
// al leerla de la base de datos. No aplica las reglas de creación: una reserva
// que fue válida cuando se hizo sigue existiendo aunque las reglas cambien.
func ReconstruirReserva(id, socioID, pistaID string, inicio time.Time, estado Estado) Reserva {
	return Reserva{
		id:      id,
		socioID: socioID,
		pistaID: pistaID,
		franja:  Franja{inicio: inicio},
		estado:  estado,
	}
}

func (r Reserva) ID() string           { return r.id }
func (r Reserva) SocioID() string      { return r.socioID }
func (r Reserva) PistaID() string      { return r.pistaID }
func (r Reserva) Franja() Franja       { return r.franja }
func (r Reserva) Estado() Estado       { return r.estado }
func (r Reserva) EstaConfirmada() bool { return r.estado == Confirmada }
