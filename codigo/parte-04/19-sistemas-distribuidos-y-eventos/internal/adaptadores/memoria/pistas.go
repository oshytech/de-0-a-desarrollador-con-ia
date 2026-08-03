package memoria

import (
	"context"
	"sync"

	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

type RepositorioPistas struct {
	mu     sync.RWMutex
	pistas map[string]dominio.Pista
}

var _ aplicacion.RepositorioPistas = (*RepositorioPistas)(nil)

func NuevoRepositorioPistas(pistas ...dominio.Pista) *RepositorioPistas {
	r := &RepositorioPistas{pistas: map[string]dominio.Pista{}}
	for _, p := range pistas {
		r.pistas[p.ID] = p
	}
	return r
}

func (r *RepositorioPistas) BuscarPorID(_ context.Context, id string) (dominio.Pista, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pista, ok := r.pistas[id]
	if !ok {
		return dominio.Pista{}, aplicacion.ErrNoEncontrado
	}
	return pista, nil
}
