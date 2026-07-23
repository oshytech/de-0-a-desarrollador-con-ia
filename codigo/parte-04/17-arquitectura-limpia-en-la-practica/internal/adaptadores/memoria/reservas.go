package memoria

import (
	"context"
	"sync"

	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

type RepositorioReservas struct {
	mu       sync.RWMutex
	reservas map[string]dominio.Reserva
}

var _ aplicacion.RepositorioReservas = (*RepositorioReservas)(nil)

func NuevoRepositorioReservas() *RepositorioReservas {
	return &RepositorioReservas{reservas: map[string]dominio.Reserva{}}
}

// Guardar no forma parte del puerto de este capítulo (ningún caso de uso lo
// necesita todavía), pero el tipo concreto puede tener más métodos que el puerto.
func (r *RepositorioReservas) Guardar(_ context.Context, reserva dominio.Reserva) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reservas[reserva.ID()] = reserva
	return nil
}

func (r *RepositorioReservas) DeSocio(_ context.Context, socioID string) ([]dominio.Reserva, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var resultado []dominio.Reserva
	for _, reserva := range r.reservas {
		if reserva.SocioID() == socioID {
			resultado = append(resultado, reserva)
		}
	}
	return resultado, nil
}
