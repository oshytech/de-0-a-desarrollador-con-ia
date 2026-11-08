package aplicacion_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"reservas/internal/adaptadores/memoria"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

func TestListarPistasOrdenadasPorID(t *testing.T) {
	pistas := memoria.NuevoRepositorioPistas(
		dominio.Pista{ID: "p-2", Nombre: "Pista 2", Deporte: dominio.Padel},
		dominio.Pista{ID: "p-1", Nombre: "Pista 1", Deporte: dominio.Tenis},
	)
	resultado, err := aplicacion.NuevoListarPistas(pistas).Ejecutar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resultado) != 2 || resultado[0].ID != "p-1" || resultado[0].Deporte != "tenis" {
		t.Fatalf("esperaba p-1 y p-2 en ese orden, salió %+v", resultado)
	}
}

// horas resume la disponibilidad como horas del reloj: [10 11 13 ...].
func horas(d aplicacion.Disponibilidad) []int {
	resultado := []int{}
	for _, inicio := range d.HorasLibres {
		resultado = append(resultado, inicio.Hour())
	}
	return resultado
}

func TestDisponibilidad(t *testing.T) {
	ctx := context.Background()
	_, pistas, reservas := memoria.DatosDeEjemplo()
	reservas.Guardar(ctx, dominio.ReconstruirReserva("r-1", "s-1", "p-1", a(1, 12), dominio.Confirmada))
	reservas.Guardar(ctx, dominio.ReconstruirReserva("r-2", "s-3", "p-1", a(1, 14), dominio.Cancelada))
	reservas.Guardar(ctx, dominio.ReconstruirReserva("r-3", "s-3", "p-2", a(1, 15), dominio.Confirmada))
	consultar := aplicacion.NuevoConsultarDisponibilidad(pistas, reservas, reloj.Fijo{Momento: ahora})

	casos := []struct {
		nombre string
		dia    int
		quiero []int
	}{
		// Son las 9:30 del día 1: las 8 y las 9 ya pasaron, las 12 están
		// ocupadas y las 14 se liberaron al cancelar. La reserva de p-2 no cuenta.
		{"hoy", 1, []int{10, 11, 13, 14, 15, 16, 17, 18, 19, 20, 21}},
		{"mañana", 2, []int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}},
		// El día 15 a las 9:30 se cumplen los 14 días de antelación.
		{"último día del plazo", 15, []int{8, 9}},
		{"fuera de plazo", 16, []int{}},
		{"ayer", 0, []int{}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := consultar.Ejecutar(ctx, "p-1", a(c.dia, 0))
			if err != nil {
				t.Fatal(err)
			}
			if got := horas(d); !slices.Equal(got, c.quiero) {
				t.Fatalf("esperaba %v, salió %v", c.quiero, got)
			}
		})
	}
}

func TestDisponibilidadDePistaDesconocida(t *testing.T) {
	_, pistas, reservas := memoria.DatosDeEjemplo()
	consultar := aplicacion.NuevoConsultarDisponibilidad(pistas, reservas, reloj.Fijo{Momento: ahora})
	if _, err := consultar.Ejecutar(context.Background(), "p-9", a(2, 0)); !errors.Is(err, aplicacion.ErrNoEncontrado) {
		t.Fatalf("esperaba ErrNoEncontrado, salió %v", err)
	}
}
