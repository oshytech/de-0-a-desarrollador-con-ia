package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"reservas/internal/dominio"
)

// Repositorio guarda y consulta reservas en PostgreSQL.
type Repositorio struct {
	pool *pgxpool.Pool
}

func NuevoRepositorio(pool *pgxpool.Pool) *Repositorio {
	return &Repositorio{pool: pool}
}

// Reservar crea una reserva confirmada de una hora dentro de una transacción.
// Las reglas que dependen de otras filas (socio activo, máximo 2 reservas
// futuras) se comprueban con la fila del socio bloqueada. El solape lo
// impide la propia base de datos con la restricción reserva_sin_solape.
func (r *Repositorio) Reservar(ctx context.Context, socioID, pistaID int64, inicio time.Time) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) // no hace nada si la transacción ya se confirmó

	var activo bool
	err = tx.QueryRow(ctx,
		"SELECT activo FROM socio WHERE id = $1 FOR UPDATE", socioID).Scan(&activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, dominio.ErrSocioNoExiste
	}
	if err != nil {
		return 0, err
	}
	if !activo {
		return 0, dominio.ErrSocioInactivo
	}

	var futuras int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM reserva
		WHERE socio_id = $1 AND estado = 'confirmada' AND inicio > now()`, socioID).Scan(&futuras)
	if err != nil {
		return 0, err
	}
	if futuras >= 2 {
		return 0, dominio.ErrLimiteReservas
	}

	var id int64
	err = tx.QueryRow(ctx, `INSERT INTO reserva (socio_id, pista_id, inicio, fin)
		VALUES ($1, $2, $3::timestamptz, $3::timestamptz + interval '1 hour')
		RETURNING id`, socioID, pistaID, inicio).Scan(&id)
	if err != nil {
		return 0, traducir(err)
	}
	return id, tx.Commit(ctx)
}

// Cancelar marca la reserva como cancelada. Al dejar de estar confirmada,
// deja de contar para la restricción de solape y el hueco queda libre.
func (r *Repositorio) Cancelar(ctx context.Context, reservaID int64) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE reserva SET estado = 'cancelada' WHERE id = $1", reservaID)
	return err
}

// traducir convierte los errores de PostgreSQL que tienen significado de
// negocio en errores del dominio. El resto se devuelve tal cual.
func traducir(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23P01": // exclusion_violation
			return dominio.ErrPistaOcupada
		case "23514": // check_violation
			return errors.Join(dominio.ErrHorarioNoValido, err)
		}
	}
	return err
}
