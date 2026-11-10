package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/jackc/pgx/v5/pgxpool"

	"reservas/datos"
	"reservas/internal/adaptadores/postgres"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
	"reservas/migraciones"
)

// nuevaBase crea un esquema vacío solo para este test, le aplica las
// migraciones y lo borra al terminar. Sin base de datos, el test se salta.
func nuevaBase(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("RESERVAS_BD")
	if url == "" {
		t.Skip("RESERVAS_BD vacía: se salta el test de integración")
	}
	ctx := context.Background()

	sufijo := make([]byte, 4)
	_, _ = rand.Read(sufijo)
	esquema := "prueba_" + hex.EncodeToString(sufijo)

	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+esquema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = admin.Exec(context.Background(), "DROP SCHEMA "+esquema+" CASCADE") })

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = esquema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	if _, err := postgres.Migrar(ctx, pool, migraciones.Ficheros); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, datos.Ejemplo); err != nil {
		t.Fatal(err)
	}
	return pool
}

var ahora = time.Date(2026, 12, 1, 9, 30, 0, 0, time.UTC)

func a(dia, hora int) time.Time { return time.Date(2026, 12, dia, hora, 0, 0, 0, time.UTC) }

func reserva(id, socio, pista string, inicio time.Time) dominio.Reserva {
	return dominio.ReconstruirReserva(id, socio, pista, inicio, dominio.Confirmada)
}

func TestRepositorioReservas(t *testing.T) {
	pool := nuevaBase(t)
	ctx := context.Background()
	repo := postgres.NuevoRepositorioReservas(pool)

	if err := repo.Guardar(ctx, reserva("r-1", "s-1", "p-1", a(2, 10))); err != nil {
		t.Fatal(err)
	}
	leida, err := repo.BuscarPorID(ctx, "r-1")
	if err != nil || !leida.Franja().Inicio().Equal(a(2, 10)) || !leida.EstaConfirmada() {
		t.Fatalf("BuscarPorID = %+v, %v", leida, err)
	}

	t.Run("el índice único impide la doble reserva", func(t *testing.T) {
		err := repo.Guardar(ctx, reserva("r-2", "s-3", "p-1", a(2, 10)))
		if !errors.Is(err, dominio.ErrPistaOcupada) {
			t.Fatalf("esperaba ErrPistaOcupada, salió %v", err)
		}
	})

	t.Run("cancelar actualiza y libera la hora", func(t *testing.T) {
		cancelada, err := leida.Cancelar("s-1", ahora)
		if err != nil {
			t.Fatal(err)
		}
		if err := repo.Guardar(ctx, cancelada); err != nil {
			t.Fatal(err)
		}
		if err := repo.Guardar(ctx, reserva("r-3", "s-3", "p-1", a(2, 10))); err != nil {
			t.Fatalf("la hora cancelada debería estar libre: %v", err)
		}
	})

	t.Run("consultas", func(t *testing.T) {
		dePista, err := repo.ConfirmadasDePista(ctx, "p-1", a(2, 10), a(2, 11))
		if err != nil || len(dePista) != 1 || dePista[0].ID() != "r-3" {
			t.Fatalf("ConfirmadasDePista = %v, %v", dePista, err)
		}
		futuras, err := repo.ConfirmadasDeSocioDesde(ctx, "s-1", ahora)
		if err != nil || len(futuras) != 0 {
			t.Fatalf("s-1 ya no tiene reservas confirmadas: %v, %v", futuras, err)
		}
		todas, err := repo.DeSocio(ctx, "s-1")
		if err != nil || len(todas) != 1 || todas[0].Estado() != dominio.Cancelada {
			t.Fatalf("DeSocio = %v, %v", todas, err)
		}
	})

	t.Run("lo que no existe", func(t *testing.T) {
		if _, err := repo.BuscarPorID(ctx, "r-99"); !errors.Is(err, aplicacion.ErrNoEncontrado) {
			t.Fatalf("reserva: %v", err)
		}
		if _, err := postgres.NuevoRepositorioSocios(pool).BuscarPorID(ctx, "s-99"); !errors.Is(err, aplicacion.ErrNoEncontrado) {
			t.Fatalf("socio: %v", err)
		}
		if _, err := postgres.NuevoRepositorioPistas(pool).BuscarPorID(ctx, "p-99"); !errors.Is(err, aplicacion.ErrNoEncontrado) {
			t.Fatalf("pista: %v", err)
		}
	})
}

// casosDeUso monta los casos de uso sobre PostgreSQL con la hora fijada.
func casosDeUso(pool *pgxpool.Pool) (*aplicacion.ReservarPista, *aplicacion.CancelarReserva, *aplicacion.ListarReservasDeSocio) {
	socios := postgres.NuevoRepositorioSocios(pool)
	pistas := postgres.NuevoRepositorioPistas(pool)
	reservas := postgres.NuevoRepositorioReservas(pool)
	relojFijo := reloj.Fijo{Momento: ahora}
	return aplicacion.NuevoReservarPista(reservas, socios, pistas, relojFijo),
		aplicacion.NuevoCancelarReserva(reservas, relojFijo),
		aplicacion.NuevoListarReservasDeSocio(socios, reservas, relojFijo)
}

// TestReglasContraPostgres recorre las seis reglas con la base de datos real.
func TestReglasContraPostgres(t *testing.T) {
	reservar, cancelar, listar := casosDeUso(nuevaBase(t))
	ctx := context.Background()
	pedir := func(socio, pista string, inicio time.Time) (aplicacion.ReservaDTO, error) {
		return reservar.Ejecutar(ctx, aplicacion.SolicitudReserva{SocioID: socio, PistaID: pista, Inicio: inicio})
	}

	primera, err := pedir("s-1", "p-1", a(2, 10))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pedir("s-1", "p-1", a(2, 10).Add(30*time.Minute)); !errors.Is(err, dominio.ErrFueraDeHorario) {
		t.Errorf("regla 1: %v", err)
	}
	if _, err := pedir("s-3", "p-1", a(2, 10)); !errors.Is(err, dominio.ErrPistaOcupada) {
		t.Errorf("regla 2: %v", err)
	}
	if _, err := pedir("s-1", "p-1", a(3, 10)); err != nil {
		t.Fatal(err)
	}
	if _, err := pedir("s-1", "p-2", a(4, 10)); !errors.Is(err, dominio.ErrLimiteDeReservas) {
		t.Errorf("regla 3: %v", err)
	}
	if _, err := pedir("s-2", "p-2", a(4, 10)); !errors.Is(err, dominio.ErrSocioInactivo) {
		t.Errorf("regla 4: %v", err)
	}
	if _, err := pedir("s-3", "p-2", a(1, 9)); !errors.Is(err, dominio.ErrFueraDePlazo) {
		t.Errorf("regla 6: %v", err)
	}
	if _, err := cancelar.Ejecutar(ctx, aplicacion.SolicitudCancelacion{ReservaID: primera.ID, SocioID: "s-1"}); err != nil {
		t.Errorf("regla 5, con margen: %v", err)
	}
	proximas, err := listar.Ejecutar(ctx, "s-1")
	if err != nil || len(proximas) != 2 || proximas[0].Estado != "cancelada" {
		t.Errorf("listado de s-1: %+v, %v", proximas, err)
	}
}

// Ocho peticiones a la vez por la misma hora: solo una puede guardarse.
func TestPeticionesSimultaneas(t *testing.T) {
	reservar, _, _ := casosDeUso(nuevaBase(t))
	var wg sync.WaitGroup
	errores := make([]error, 8)
	for i := range errores {
		socio := []string{"s-1", "s-3"}[i%2]
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errores[i] = reservar.Ejecutar(context.Background(),
				aplicacion.SolicitudReserva{SocioID: socio, PistaID: "p-2", Inicio: a(3, 20)})
		}()
	}
	wg.Wait()
	bien := 0
	for _, err := range errores {
		switch {
		case err == nil:
			bien++
		case !errors.Is(err, dominio.ErrPistaOcupada):
			t.Fatalf("error inesperado: %v", err)
		}
	}
	if bien != 1 {
		t.Fatalf("se guardaron %d reservas para la misma hora, esperaba 1", bien)
	}
}

// Las dos consultas que añade el capítulo 47, con la hora del club en Madrid
// como en producción: las horas se guardan en UTC y se cuentan en hora local.
func TestPistasYDisponibilidadContraPostgres(t *testing.T) {
	pool := nuevaBase(t)
	ctx := context.Background()
	madrid, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		t.Fatal(err)
	}
	pistas := postgres.NuevoRepositorioPistas(pool)
	reservas := postgres.NuevoRepositorioReservas(pool)

	lista, err := aplicacion.NuevoListarPistas(pistas).Ejecutar(ctx)
	if err != nil || len(lista) != 2 || lista[0].ID != "p-1" || lista[1].Deporte != "padel" {
		t.Fatalf("ListarPistas = %+v, %v", lista, err)
	}

	enMadrid := func(hora int) time.Time { return time.Date(2026, 12, 2, hora, 0, 0, 0, madrid) }
	if err := reservas.Guardar(ctx, reserva("r-1", "s-1", "p-1", enMadrid(8))); err != nil {
		t.Fatal(err)
	}
	if err := reservas.Guardar(ctx, reserva("r-2", "s-3", "p-1", enMadrid(21))); err != nil {
		t.Fatal(err)
	}
	consultar := aplicacion.NuevoConsultarDisponibilidad(pistas, reservas, reloj.Fijo{Momento: ahora})
	d, err := consultar.Ejecutar(ctx, "p-1", enMadrid(0))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.HorasLibres) != 12 || !d.HorasLibres[0].Equal(enMadrid(9)) || !d.HorasLibres[11].Equal(enMadrid(20)) {
		t.Fatalf("esperaba de 09:00 a 20:00 en Madrid, salió %v", d.HorasLibres)
	}
	if _, err := consultar.Ejecutar(ctx, "p-99", enMadrid(0)); !errors.Is(err, aplicacion.ErrNoEncontrado) {
		t.Fatalf("pista desconocida: %v", err)
	}
}
