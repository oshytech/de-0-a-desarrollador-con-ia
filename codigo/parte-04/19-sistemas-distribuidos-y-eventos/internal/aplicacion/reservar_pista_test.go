package aplicacion_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"reservas/internal/adaptadores/memoria"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

// espiaEventos guarda lo que se publica para poder comprobarlo.
type espiaEventos struct {
	mu         sync.Mutex
	publicados []dominio.Evento
}

func (e *espiaEventos) Publicar(_ context.Context, eventos ...dominio.Evento) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.publicados = append(e.publicados, eventos...)
	return nil
}

func nuevoReservar() (*aplicacion.ReservarPista, *memoria.RepositorioReservas) {
	socios, pistas, reservas := memoria.DatosDeEjemplo()
	return aplicacion.NuevoReservarPista(reservas, socios, pistas, reloj.Fijo{Momento: ahora}, &espiaEventos{}), reservas
}

func TestReservarGuardaLaReserva(t *testing.T) {
	reservar, reservas := nuevoReservar()
	ctx := context.Background()
	dto, err := reservar.Ejecutar(ctx, aplicacion.SolicitudReserva{SocioID: "s-1", PistaID: "p-1", Inicio: a(2, 10)})
	if err != nil {
		t.Fatal(err)
	}
	guardada, err := reservas.BuscarPorID(ctx, dto.ID)
	if err != nil || !guardada.EstaConfirmada() {
		t.Fatalf("la reserva debería estar guardada y confirmada: %v", err)
	}
}

func TestLaTerceraReservaFuturaSeRechaza(t *testing.T) {
	reservar, _ := nuevoReservar()
	ctx := context.Background()
	for _, dia := range []int{2, 3} {
		if _, err := reservar.Ejecutar(ctx, aplicacion.SolicitudReserva{SocioID: "s-1", PistaID: "p-1", Inicio: a(dia, 10)}); err != nil {
			t.Fatal(err)
		}
	}
	_, err := reservar.Ejecutar(ctx, aplicacion.SolicitudReserva{SocioID: "s-1", PistaID: "p-2", Inicio: a(4, 10)})
	if !errors.Is(err, dominio.ErrLimiteDeReservas) {
		t.Fatalf("esperaba ErrLimiteDeReservas, salió %v", err)
	}
}

// Dos socios piden la misma pista a la misma hora a la vez. Los dos pueden
// pasar las comprobaciones del dominio antes de que el otro guarde, así que
// quien impide la doble reserva es el repositorio. Solo una puede salir bien.
func TestDosPeticionesSimultaneasParaLaMismaHora(t *testing.T) {
	reservar, _ := nuevoReservar()
	var wg sync.WaitGroup
	errores := make([]error, 2)
	for i, socio := range []string{"s-1", "s-3"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errores[i] = reservar.Ejecutar(context.Background(),
				aplicacion.SolicitudReserva{SocioID: socio, PistaID: "p-1", Inicio: a(2, 10)})
		}()
	}
	wg.Wait()
	bien, ocupada := 0, 0
	for _, err := range errores {
		switch {
		case err == nil:
			bien++
		case errors.Is(err, dominio.ErrPistaOcupada):
			ocupada++
		default:
			t.Fatalf("error inesperado: %v", err)
		}
	}
	if bien != 1 || ocupada != 1 {
		t.Fatalf("esperaba una reserva y un rechazo, salieron %d y %d", bien, ocupada)
	}
}

func TestCancelarConDosHorasDeMargen(t *testing.T) {
	socios, pistas, reservas := memoria.DatosDeEjemplo()
	ctx := context.Background()
	alasNueve := reloj.Fijo{Momento: time.Date(2026, 12, 2, 9, 0, 0, 0, time.UTC)}
	reservar := aplicacion.NuevoReservarPista(reservas, socios, pistas, reloj.Fijo{Momento: ahora}, &espiaEventos{})
	dto, _ := reservar.Ejecutar(ctx, aplicacion.SolicitudReserva{SocioID: "s-1", PistaID: "p-1", Inicio: a(2, 10)})

	cancelar := aplicacion.NuevoCancelarReserva(reservas, alasNueve, &espiaEventos{}) // falta una hora
	_, err := cancelar.Ejecutar(ctx, aplicacion.SolicitudCancelacion{ReservaID: dto.ID, SocioID: "s-1"})
	if !errors.Is(err, dominio.ErrCancelacionFueraDePlazo) {
		t.Fatalf("esperaba ErrCancelacionFueraDePlazo, salió %v", err)
	}
	if _, err := cancelar.Ejecutar(ctx, aplicacion.SolicitudCancelacion{ReservaID: "r-99", SocioID: "s-1"}); !errors.Is(err, aplicacion.ErrNoEncontrado) {
		t.Fatalf("esperaba ErrNoEncontrado, salió %v", err)
	}
}
