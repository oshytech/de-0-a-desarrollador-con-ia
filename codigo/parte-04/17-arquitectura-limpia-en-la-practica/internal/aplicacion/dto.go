package aplicacion

import (
	"time"

	"reservas/internal/dominio"
)

// ReservaDTO es lo que los casos de uso devuelven hacia fuera. No lleva
// etiquetas JSON: cómo se serializa lo decide cada adaptador.
type ReservaDTO struct {
	ID      string
	SocioID string
	PistaID string
	Inicio  time.Time
	Fin     time.Time
	Estado  string
}

func aDTO(r dominio.Reserva) ReservaDTO {
	return ReservaDTO{
		ID:      r.ID(),
		SocioID: r.SocioID(),
		PistaID: r.PistaID(),
		Inicio:  r.Franja().Inicio(),
		Fin:     r.Franja().Fin(),
		Estado:  string(r.Estado()),
	}
}
