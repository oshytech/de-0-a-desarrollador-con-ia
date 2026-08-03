package eventos

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"reservas/internal/dominio"
)

var confirmada = dominio.ReservaConfirmada{ReservaID: "r-1", SocioID: "s-1", PistaID: "p-1",
	Inicio: time.Date(2026, 12, 2, 10, 0, 0, 0, time.UTC)}

func TestCadaSuscriptorRecibeSuCopia(t *testing.T) {
	bus := NuevoBus(1, 0)
	var mu sync.Mutex
	recibidos := map[string]int{}
	for _, nombre := range []string{"avisos", "estadisticas"} {
		bus.Suscribir(nombre, func(_ context.Context, e dominio.Evento) error {
			mu.Lock()
			defer mu.Unlock()
			recibidos[nombre]++
			return nil
		})
	}
	if err := bus.Publicar(context.Background(), confirmada); err != nil {
		t.Fatal(err)
	}
	bus.Cerrar() // espera a que terminen
	if recibidos["avisos"] != 1 || recibidos["estadisticas"] != 1 {
		t.Fatalf("cada suscriptor debería recibir el evento una vez: %v", recibidos)
	}
}

func TestReintentaYDespuesLoApartaComoFallido(t *testing.T) {
	bus := NuevoBus(3, time.Millisecond)
	intentos := 0
	bus.Suscribir("siempre-falla", func(context.Context, dominio.Evento) error {
		intentos++
		return errors.New("servicio de correo caído")
	})
	bus.Publicar(context.Background(), confirmada)
	bus.Cerrar()

	if intentos != 3 {
		t.Fatalf("esperaba 3 intentos, hubo %d", intentos)
	}
	if fallidos := bus.Fallidos(); len(fallidos) != 1 || fallidos[0].Suscriptor != "siempre-falla" {
		t.Fatalf("el evento debería quedar apartado: %+v", fallidos)
	}
}

func TestUnFalloTransitorioSeRecupera(t *testing.T) {
	bus := NuevoBus(3, time.Millisecond)
	intentos := 0
	bus.Suscribir("falla-una-vez", func(context.Context, dominio.Evento) error {
		intentos++
		if intentos == 1 {
			return errors.New("tiempo de espera agotado")
		}
		return nil
	})
	bus.Publicar(context.Background(), confirmada)
	bus.Cerrar()
	if intentos != 2 || len(bus.Fallidos()) != 0 {
		t.Fatalf("esperaba éxito al segundo intento: %d intentos, %d fallidos", intentos, len(bus.Fallidos()))
	}
}

func TestNoSePublicaConElBusCerrado(t *testing.T) {
	bus := NuevoBus(1, 0)
	bus.Cerrar()
	if err := bus.Publicar(context.Background(), confirmada); !errors.Is(err, ErrBusCerrado) {
		t.Fatalf("esperaba ErrBusCerrado, salió %v", err)
	}
}
