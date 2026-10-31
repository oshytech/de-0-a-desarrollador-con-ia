package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Test de integración. Necesita una base de datos PostgreSQL de usar y tirar:
//
//	docker run --rm -d --name club-pg -e POSTGRES_PASSWORD=solo-pruebas -p 55432:5432 postgres:17-alpine
//	RESERVAS_TEST_URL_BD=postgres://postgres:solo-pruebas@localhost:55432/postgres go test ./...
//
// Sin la variable, el test se salta.
func conectar(t *testing.T) *pgx.Conn {
	t.Helper()
	url := os.Getenv("RESERVAS_TEST_URL_BD")
	if url == "" {
		t.Skip("RESERVAS_TEST_URL_BD no definida: se salta el test de integración")
	}
	ctx := context.Background()
	db, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close(ctx) })

	_, err = db.Exec(ctx, `
		DROP TABLE IF EXISTS reservas;
		CREATE TABLE reservas (
			id       bigint PRIMARY KEY,
			socio_id bigint NOT NULL,
			pista_id text NOT NULL,
			inicio   timestamptz NOT NULL,
			estado   text NOT NULL
		);
		INSERT INTO reservas VALUES
			(1043, 1, 'padel-2', '2026-12-01 18:00+01', 'confirmada'),
			(1044, 2, 'tenis-1', '2026-12-01 19:00+01', 'confirmada'),
			(1045, 3, 'padel_9', '2026-12-02 09:00+01', 'confirmada');`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

const inyeccion = "' OR 1=1 --"

func TestVulnerable_LaInyeccionDevuelveReservasAjenas(t *testing.T) {
	db := conectar(t)
	reservas, err := BuscarReservasDeSocioVulnerable(context.Background(), db, 1, inyeccion)
	if err != nil {
		t.Fatal(err)
	}
	if len(reservas) != 3 {
		t.Fatalf("esperaba demostrar la fuga (3 reservas), obtuve %d", len(reservas))
	}
}

func TestParametrizada_ElFiltroEsSoloUnDato(t *testing.T) {
	db := conectar(t)
	ctx := context.Background()

	reservas, err := BuscarReservasDeSocio(ctx, db, 1, inyeccion)
	if err != nil {
		t.Fatal(err)
	}
	if len(reservas) != 0 {
		t.Fatalf("la inyección no debería devolver nada, devolvió %d reservas", len(reservas))
	}

	reservas, err = BuscarReservasDeSocio(ctx, db, 1, "padel")
	if err != nil {
		t.Fatal(err)
	}
	if len(reservas) != 1 || reservas[0].ID != 1043 {
		t.Fatalf("la búsqueda normal debería devolver solo la 1043: %+v", reservas)
	}
}

func TestParametrizada_ComodinesLiterales(t *testing.T) {
	db := conectar(t)
	// "_" en LIKE significa "un carácter cualquiera". Escapado, solo
	// encuentra pistas que contienen un guion bajo de verdad.
	reservas, err := BuscarReservasDeSocio(context.Background(), db, 3, "_")
	if err != nil {
		t.Fatal(err)
	}
	if len(reservas) != 1 {
		t.Fatalf("esperaba 1 reserva, obtuve %d", len(reservas))
	}
	reservas, err = BuscarReservasDeSocio(context.Background(), db, 1, "_")
	if err != nil {
		t.Fatal(err)
	}
	if len(reservas) != 0 {
		t.Fatalf("padel-2 no contiene _ y no debería aparecer: %+v", reservas)
	}
}
