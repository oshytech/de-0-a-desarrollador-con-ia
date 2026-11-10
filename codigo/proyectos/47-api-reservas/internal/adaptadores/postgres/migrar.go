// El aplicador de migraciones es el del capítulo 37, sin cambios salvo el
// número del cerrojo.

package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// cerrojoMigraciones identifica el bloqueo consultivo que impide que dos
// instancias de la aplicación migren a la vez.
const cerrojoMigraciones = 4701

// Migrar aplica, en orden y una sola vez, los ficheros .sql que aún no
// constan en la tabla migracion_aplicada. Devuelve los que ha aplicado.
func Migrar(ctx context.Context, pool *pgxpool.Pool, ficheros fs.FS) ([]string, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", cerrojoMigraciones); err != nil {
		return nil, err
	}
	defer conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", cerrojoMigraciones)

	_, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS migracion_aplicada (
		version     text PRIMARY KEY,
		aplicada_en timestamptz NOT NULL DEFAULT now())`)
	if err != nil {
		return nil, err
	}

	nombres, err := fs.Glob(ficheros, "*.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(nombres) // 001_..., 002_..., 003_...

	var aplicadas []string
	for _, nombre := range nombres {
		var ya bool
		err := conn.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM migracion_aplicada WHERE version = $1)", nombre).Scan(&ya)
		if err != nil {
			return aplicadas, err
		}
		if ya {
			continue
		}
		sql, err := fs.ReadFile(ficheros, nombre)
		if err != nil {
			return aplicadas, err
		}
		// Cada migración va en su transacción: o se aplica entera y queda
		// registrada, o no se aplica nada.
		tx, err := conn.Begin(ctx)
		if err != nil {
			return aplicadas, err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			tx.Rollback(ctx)
			return aplicadas, fmt.Errorf("migración %s: %w", nombre, err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO migracion_aplicada (version) VALUES ($1)", nombre); err != nil {
			tx.Rollback(ctx)
			return aplicadas, err
		}
		if err := tx.Commit(ctx); err != nil {
			return aplicadas, err
		}
		aplicadas = append(aplicadas, nombre)
	}
	return aplicadas, nil
}
