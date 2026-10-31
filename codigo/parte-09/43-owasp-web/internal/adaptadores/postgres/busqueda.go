// Package postgres: consultas de reservas contra PostgreSQL.
package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"owaspweb/internal/dominio"
)

// Consultor es la parte de *pgx.Conn (o de un pool) que usamos.
type Consultor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

const columnas = "id, socio_id, pista_id, inicio, estado"

// BuscarReservasDeSocioVulnerable mete el filtro del usuario dentro del SQL.
// Con el filtro  ' OR 1=1 --  la condición del socio deja de importar.
// Solo existe para que el test de integración demuestre el problema.
func BuscarReservasDeSocioVulnerable(ctx context.Context, db Consultor, socioID int64, filtroPista string) ([]dominio.Reserva, error) {
	sql := fmt.Sprintf(
		"SELECT %s FROM reservas WHERE socio_id = %d AND pista_id LIKE '%%%s%%' ORDER BY id",
		columnas, socioID, filtroPista)
	return consultar(ctx, db, sql)
}

// BuscarReservasDeSocio envía la consulta y los valores por separado. El
// filtro viaja como dato en $2 y PostgreSQL nunca lo interpreta como SQL.
// Además se escapan % y _ para que el usuario busque texto literal.
func BuscarReservasDeSocio(ctx context.Context, db Consultor, socioID int64, filtroPista string) ([]dominio.Reserva, error) {
	const sql = "SELECT " + columnas + " FROM reservas " +
		`WHERE socio_id = $1 AND pista_id LIKE '%' || $2 || '%' ESCAPE '\' ORDER BY id`
	return consultar(ctx, db, sql, socioID, escaparLike(filtroPista))
}

func escaparLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func consultar(ctx context.Context, db Consultor, sql string, args ...any) ([]dominio.Reserva, error) {
	filas, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(filas, func(f pgx.CollectableRow) (dominio.Reserva, error) {
		var r dominio.Reserva
		var estado string
		err := f.Scan(&r.ID, &r.SocioID, &r.PistaID, &r.Inicio, &estado)
		r.Estado = dominio.Estado(estado)
		return r, err
	})
}
