// Package postgres implementa los repositorios con PostgreSQL y el driver pgx.
package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

// Comprobaciones en tiempo de compilación: si un puerto cambia, esto no compila.
var (
	_ aplicacion.RepositorioSocios   = (*RepositorioSocios)(nil)
	_ aplicacion.RepositorioPistas   = (*RepositorioPistas)(nil)
	_ aplicacion.RepositorioReservas = (*RepositorioReservas)(nil)
)

// indiceSinSolapes es el índice único de migraciones/001_esquema.sql
// que impide dos reservas confirmadas de la misma pista a la misma hora.
const indiceSinSolapes = "reserva_pista_inicio_confirmada"

type RepositorioSocios struct{ pool *pgxpool.Pool }

func NuevoRepositorioSocios(pool *pgxpool.Pool) *RepositorioSocios {
	return &RepositorioSocios{pool: pool}
}

func (r *RepositorioSocios) BuscarPorID(ctx context.Context, id string) (dominio.Socio, error) {
	var s dominio.Socio
	err := r.pool.QueryRow(ctx,
		`SELECT id, nombre, correo, activo FROM socio WHERE id = $1`, id).
		Scan(&s.ID, &s.Nombre, &s.Correo, &s.Activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return dominio.Socio{}, aplicacion.ErrNoEncontrado
	}
	return s, err
}

type RepositorioPistas struct{ pool *pgxpool.Pool }

func NuevoRepositorioPistas(pool *pgxpool.Pool) *RepositorioPistas {
	return &RepositorioPistas{pool: pool}
}

func (r *RepositorioPistas) BuscarPorID(ctx context.Context, id string) (dominio.Pista, error) {
	var p dominio.Pista
	var deporte string
	err := r.pool.QueryRow(ctx,
		`SELECT id, nombre, deporte FROM pista WHERE id = $1`, id).
		Scan(&p.ID, &p.Nombre, &deporte)
	if errors.Is(err, pgx.ErrNoRows) {
		return dominio.Pista{}, aplicacion.ErrNoEncontrado
	}
	p.Deporte = dominio.Deporte(deporte)
	return p, err
}

func (r *RepositorioPistas) Todas(ctx context.Context) ([]dominio.Pista, error) {
	filas, err := r.pool.Query(ctx, `SELECT id, nombre, deporte FROM pista ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(filas, func(fila pgx.CollectableRow) (dominio.Pista, error) {
		var p dominio.Pista
		var deporte string
		err := fila.Scan(&p.ID, &p.Nombre, &deporte)
		p.Deporte = dominio.Deporte(deporte)
		return p, err
	})
}

type RepositorioReservas struct{ pool *pgxpool.Pool }

func NuevoRepositorioReservas(pool *pgxpool.Pool) *RepositorioReservas {
	return &RepositorioReservas{pool: pool}
}

// NuevoID genera un identificador aleatorio. No consulta la base de datos,
// así que no puede fallar ni necesita contexto.
func (r *RepositorioReservas) NuevoID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // crypto/rand no devuelve error en las plataformas soportadas
	return "r-" + hex.EncodeToString(b)
}

// Guardar crea la reserva o, si ya existe, actualiza su estado. El solape
// lo vigila el índice único: si dos peticiones llegan a la vez, la segunda
// en escribir recibe dominio.ErrPistaOcupada.
func (r *RepositorioReservas) Guardar(ctx context.Context, reserva dominio.Reserva) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO reserva (id, socio_id, pista_id, inicio, fin, estado)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO UPDATE SET estado = EXCLUDED.estado`,
		reserva.ID(), reserva.SocioID(), reserva.PistaID(),
		reserva.Franja().Inicio(), reserva.Franja().Fin(), string(reserva.Estado()))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == indiceSinSolapes {
		return dominio.ErrPistaOcupada
	}
	if err != nil {
		return fmt.Errorf("guardar reserva %s: %w", reserva.ID(), err)
	}
	return nil
}

const columnas = `id, socio_id, pista_id, inicio, estado`

func (r *RepositorioReservas) BuscarPorID(ctx context.Context, id string) (dominio.Reserva, error) {
	filas, err := r.pool.Query(ctx, `SELECT `+columnas+` FROM reserva WHERE id = $1`, id)
	if err != nil {
		return dominio.Reserva{}, err
	}
	reserva, err := pgx.CollectExactlyOneRow(filas, aReserva)
	if errors.Is(err, pgx.ErrNoRows) {
		return dominio.Reserva{}, aplicacion.ErrNoEncontrado
	}
	return reserva, err
}

// ConfirmadasDePista devuelve las reservas confirmadas que se solapan con [desde, hasta).
func (r *RepositorioReservas) ConfirmadasDePista(ctx context.Context, pistaID string, desde, hasta time.Time) ([]dominio.Reserva, error) {
	return r.consultar(ctx,
		`SELECT `+columnas+` FROM reserva
		 WHERE pista_id = $1 AND estado = 'confirmada' AND inicio < $3 AND fin > $2`,
		pistaID, desde, hasta)
}

func (r *RepositorioReservas) ConfirmadasDeSocioDesde(ctx context.Context, socioID string, desde time.Time) ([]dominio.Reserva, error) {
	return r.consultar(ctx,
		`SELECT `+columnas+` FROM reserva
		 WHERE socio_id = $1 AND estado = 'confirmada' AND inicio > $2`,
		socioID, desde)
}

func (r *RepositorioReservas) DeSocio(ctx context.Context, socioID string) ([]dominio.Reserva, error) {
	return r.consultar(ctx,
		`SELECT `+columnas+` FROM reserva WHERE socio_id = $1 ORDER BY inicio`, socioID)
}

func (r *RepositorioReservas) consultar(ctx context.Context, sql string, argumentos ...any) ([]dominio.Reserva, error) {
	filas, err := r.pool.Query(ctx, sql, argumentos...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(filas, aReserva)
}

// aReserva convierte una fila en una reserva. Usa ReconstruirReserva porque
// una reserva guardada no tiene que volver a pasar las reglas de creación.
func aReserva(fila pgx.CollectableRow) (dominio.Reserva, error) {
	var id, socioID, pistaID, estado string
	var inicio time.Time
	if err := fila.Scan(&id, &socioID, &pistaID, &inicio, &estado); err != nil {
		return dominio.Reserva{}, err
	}
	return dominio.ReconstruirReserva(id, socioID, pistaID, inicio, dominio.Estado(estado)), nil
}
