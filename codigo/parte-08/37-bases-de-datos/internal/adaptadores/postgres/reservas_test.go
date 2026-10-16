package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/jackc/pgx/v5/pgxpool"

	"reservas/internal/adaptadores/postgres"
	"reservas/internal/dominio"
	"reservas/migraciones"
)

// Estos tests necesitan un PostgreSQL de verdad. Si RESERVAS_BD no está
// definida, se saltan (así "go test ./..." funciona en cualquier sitio).
func conectar(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("RESERVAS_BD")
	if url == "" {
		t.Skip("RESERVAS_BD no definida: test de integración omitido")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := postgres.Migrar(ctx, pool, migraciones.Ficheros); err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `
		TRUNCATE reserva;
		INSERT INTO socio (nombre, correo, activo) VALUES
			('Test 1', 't1@ejemplo.org', true), ('Test 2', 't2@ejemplo.org', true),
			('Test 3', 't3@ejemplo.org', false)
		ON CONFLICT (correo) DO NOTHING;
		INSERT INTO pista (nombre, deporte) VALUES ('Pista test', 'tenis')
		ON CONFLICT (nombre) DO NOTHING;`)
	if err != nil {
		t.Fatal(err)
	}
	return pool
}

func id(t *testing.T, pool *pgxpool.Pool, consulta, valor string) int64 {
	t.Helper()
	var n int64
	if err := pool.QueryRow(context.Background(), consulta, valor).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func mananaALas(hora int) time.Time {
	madrid, _ := time.LoadLocation("Europe/Madrid")
	m := time.Now().In(madrid).AddDate(0, 0, 1)
	return time.Date(m.Year(), m.Month(), m.Day(), hora, 0, 0, 0, madrid)
}

func TestReglasEnLaBaseDeDatos(t *testing.T) {
	pool := conectar(t)
	ctx := context.Background()
	repo := postgres.NuevoRepositorio(pool)
	s1 := id(t, pool, "SELECT id FROM socio WHERE correo = $1", "t1@ejemplo.org")
	s2 := id(t, pool, "SELECT id FROM socio WHERE correo = $1", "t2@ejemplo.org")
	s3 := id(t, pool, "SELECT id FROM socio WHERE correo = $1", "t3@ejemplo.org")
	pista := id(t, pool, "SELECT id FROM pista WHERE nombre = $1", "Pista test")

	primera, err := repo.Reservar(ctx, s1, pista, mananaALas(10))
	if err != nil {
		t.Fatalf("la primera reserva debía entrar: %v", err)
	}

	t.Run("solape en la misma pista", func(t *testing.T) {
		_, err := repo.Reservar(ctx, s2, pista, mananaALas(10))
		if !errors.Is(err, dominio.ErrPistaOcupada) {
			t.Fatalf("esperaba ErrPistaOcupada, llegó %v", err)
		}
	})

	t.Run("cancelar libera el hueco", func(t *testing.T) {
		if err := repo.Cancelar(ctx, primera); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.Reservar(ctx, s2, pista, mananaALas(10)); err != nil {
			t.Fatalf("el hueco debía estar libre: %v", err)
		}
	})

	t.Run("fuera de horario", func(t *testing.T) {
		_, err := repo.Reservar(ctx, s1, pista, mananaALas(7))
		if !errors.Is(err, dominio.ErrHorarioNoValido) {
			t.Fatalf("esperaba ErrHorarioNoValido, llegó %v", err)
		}
	})

	t.Run("socio inactivo", func(t *testing.T) {
		_, err := repo.Reservar(ctx, s3, pista, mananaALas(16))
		if !errors.Is(err, dominio.ErrSocioInactivo) {
			t.Fatalf("esperaba ErrSocioInactivo, llegó %v", err)
		}
	})

	t.Run("máximo dos reservas futuras", func(t *testing.T) {
		for _, h := range []int{17, 18} {
			if _, err := repo.Reservar(ctx, s1, pista, mananaALas(h)); err != nil {
				t.Fatal(err)
			}
		}
		_, err := repo.Reservar(ctx, s1, pista, mananaALas(19))
		if !errors.Is(err, dominio.ErrLimiteReservas) {
			t.Fatalf("esperaba ErrLimiteReservas, llegó %v", err)
		}
	})
}
