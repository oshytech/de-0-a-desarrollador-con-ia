package memoria

import (
	"context"
	"fmt"
	"sync"
	"time"

	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

type RepositorioReservas struct {
	mu        sync.RWMutex
	reservas  map[string]dominio.Reserva
	siguiente int
}

var _ aplicacion.RepositorioReservas = (*RepositorioReservas)(nil)

func NuevoRepositorioReservas() *RepositorioReservas {
	return &RepositorioReservas{reservas: map[string]dominio.Reserva{}}
}

func (r *RepositorioReservas) NuevoID() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.siguiente++
	return fmt.Sprintf("r-%d", r.siguiente)
}

// Guardar comprueba el solape dentro del cerrojo. Es lo que en PostgreSQL
// haría un índice único sobre (pista, inicio) para las reservas confirmadas.
func (r *RepositorioReservas) Guardar(_ context.Context, nueva dominio.Reserva) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if nueva.EstaConfirmada() {
		for id, otra := range r.reservas {
			if id != nueva.ID() && otra.EstaConfirmada() && otra.PistaID() == nueva.PistaID() &&
				otra.Franja().SeSolapaCon(nueva.Franja()) {
				return dominio.ErrPistaOcupada
			}
		}
	}
	r.reservas[nueva.ID()] = nueva
	return nil
}

func (r *RepositorioReservas) BuscarPorID(_ context.Context, id string) (dominio.Reserva, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reserva, ok := r.reservas[id]
	if !ok {
		return dominio.Reserva{}, aplicacion.ErrNoEncontrado
	}
	return reserva, nil
}

func (r *RepositorioReservas) ConfirmadasDePista(_ context.Context, pistaID string, desde, hasta time.Time) ([]dominio.Reserva, error) {
	return r.filtrar(func(reserva dominio.Reserva) bool {
		f := reserva.Franja()
		return reserva.PistaID() == pistaID && reserva.EstaConfirmada() &&
			f.Inicio().Before(hasta) && f.Fin().After(desde)
	}), nil
}

func (r *RepositorioReservas) ConfirmadasDeSocioDesde(_ context.Context, socioID string, desde time.Time) ([]dominio.Reserva, error) {
	return r.filtrar(func(reserva dominio.Reserva) bool {
		return reserva.SocioID() == socioID && reserva.EstaConfirmada() && reserva.Franja().Inicio().After(desde)
	}), nil
}

func (r *RepositorioReservas) DeSocio(_ context.Context, socioID string) ([]dominio.Reserva, error) {
	return r.filtrar(func(reserva dominio.Reserva) bool { return reserva.SocioID() == socioID }), nil
}

func (r *RepositorioReservas) filtrar(cumple func(dominio.Reserva) bool) []dominio.Reserva {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var resultado []dominio.Reserva
	for _, reserva := range r.reservas {
		if cumple(reserva) {
			resultado = append(resultado, reserva)
		}
	}
	return resultado
}
