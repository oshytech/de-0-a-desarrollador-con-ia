// Package memoria implementa los repositorios en memoria, para tests y demos.
package memoria

import (
	"context"
	"sync"

	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

type RepositorioSocios struct {
	mu     sync.RWMutex
	socios map[string]dominio.Socio
}

// Comprobación en tiempo de compilación: si el puerto cambia, esto no compila.
var _ aplicacion.RepositorioSocios = (*RepositorioSocios)(nil)

func NuevoRepositorioSocios(socios ...dominio.Socio) *RepositorioSocios {
	r := &RepositorioSocios{socios: map[string]dominio.Socio{}}
	for _, s := range socios {
		r.socios[s.ID] = s
	}
	return r
}

func (r *RepositorioSocios) BuscarPorID(_ context.Context, id string) (dominio.Socio, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	socio, ok := r.socios[id]
	if !ok {
		return dominio.Socio{}, aplicacion.ErrNoEncontrado
	}
	return socio, nil
}
