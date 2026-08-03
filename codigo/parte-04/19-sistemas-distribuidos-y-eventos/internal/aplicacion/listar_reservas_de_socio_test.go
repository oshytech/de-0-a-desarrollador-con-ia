package aplicacion_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"reservas/internal/adaptadores/memoria"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

var ahora = time.Date(2026, 12, 1, 9, 30, 0, 0, time.UTC)

func a(dia, hora int) time.Time { return time.Date(2026, 12, dia, hora, 0, 0, 0, time.UTC) }

func TestListaSoloLasProximasReservasOrdenadas(t *testing.T) {
	ctx := context.Background()
	socios := memoria.NuevoRepositorioSocios(dominio.Socio{ID: "s-1", Nombre: "Ana", Activo: true})
	reservas := memoria.NuevoRepositorioReservas()
	reservas.Guardar(ctx, dominio.ReconstruirReserva("r-1", "s-1", "p-1", a(3, 18), dominio.Confirmada))
	reservas.Guardar(ctx, dominio.ReconstruirReserva("r-2", "s-1", "p-2", a(1, 8), dominio.Confirmada)) // ya pasó
	reservas.Guardar(ctx, dominio.ReconstruirReserva("r-3", "s-1", "p-1", a(2, 10), dominio.Cancelada))

	listar := aplicacion.NuevoListarReservasDeSocio(socios, reservas, reloj.Fijo{Momento: ahora})
	resultado, err := listar.Ejecutar(ctx, "s-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(resultado) != 2 || resultado[0].ID != "r-3" || resultado[1].ID != "r-1" {
		t.Fatalf("esperaba r-3 y r-1 en ese orden, salió %+v", resultado)
	}
}

func TestSocioDesconocido(t *testing.T) {
	listar := aplicacion.NuevoListarReservasDeSocio(
		memoria.NuevoRepositorioSocios(), memoria.NuevoRepositorioReservas(), reloj.Fijo{Momento: ahora})
	_, err := listar.Ejecutar(context.Background(), "s-99")
	if !errors.Is(err, aplicacion.ErrNoEncontrado) {
		t.Fatalf("esperaba ErrNoEncontrado, salió %v", err)
	}
}
