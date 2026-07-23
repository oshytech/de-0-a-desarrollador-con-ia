package aplicacion

import (
	"context"
	"errors"
	"time"

	"reservas/internal/dominio"
)

// ErrNoEncontrado lo devuelven los repositorios cuando no existe lo que se busca.
var ErrNoEncontrado = errors.New("no encontrado")

// Los puertos se definen aquí, en el lado de quien los necesita,
// y solo con las operaciones que los casos de uso usan de verdad.

type RepositorioSocios interface {
	BuscarPorID(ctx context.Context, id string) (dominio.Socio, error)
}

type RepositorioReservas interface {
	DeSocio(ctx context.Context, socioID string) ([]dominio.Reserva, error)
}

type Reloj interface {
	Ahora() time.Time
}
