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

type RepositorioPistas interface {
	BuscarPorID(ctx context.Context, id string) (dominio.Pista, error)
	// Todas devuelve las pistas del club, en cualquier orden. La añade el
	// proyecto del capítulo 47 para el caso de uso ListarPistas.
	Todas(ctx context.Context) ([]dominio.Pista, error)
}

type RepositorioReservas interface {
	// NuevoID da un identificador para una reserva que aún no existe.
	NuevoID() string
	// Guardar crea o actualiza. Si otra reserva confirmada ocupa ya la misma
	// pista y hora, devuelve dominio.ErrPistaOcupada: es la última defensa
	// cuando dos peticiones llegan a la vez.
	Guardar(ctx context.Context, r dominio.Reserva) error
	BuscarPorID(ctx context.Context, id string) (dominio.Reserva, error)
	ConfirmadasDePista(ctx context.Context, pistaID string, desde, hasta time.Time) ([]dominio.Reserva, error)
	ConfirmadasDeSocioDesde(ctx context.Context, socioID string, desde time.Time) ([]dominio.Reserva, error)
	DeSocio(ctx context.Context, socioID string) ([]dominio.Reserva, error)
}

type Reloj interface {
	Ahora() time.Time
}
