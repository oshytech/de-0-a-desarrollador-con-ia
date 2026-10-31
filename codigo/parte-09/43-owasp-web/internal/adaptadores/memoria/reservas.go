// Package memoria: repositorio de reservas en memoria para tests y demos.
package memoria

import (
	"context"
	"sync"

	"owaspweb/internal/dominio"
)

type Reservas struct {
	mu    sync.Mutex
	datos map[int64]dominio.Reserva
}

func NuevasReservas(iniciales ...dominio.Reserva) *Reservas {
	r := &Reservas{datos: map[int64]dominio.Reserva{}}
	for _, x := range iniciales {
		r.datos[x.ID] = x
	}
	return r
}

func (r *Reservas) Buscar(_ context.Context, id int64) (dominio.Reserva, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	x, ok := r.datos[id]
	if !ok {
		return dominio.Reserva{}, dominio.ErrReservaNoEncontrada
	}
	return x, nil
}

func (r *Reservas) Guardar(_ context.Context, x dominio.Reserva) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.datos[x.ID] = x
	return nil
}
