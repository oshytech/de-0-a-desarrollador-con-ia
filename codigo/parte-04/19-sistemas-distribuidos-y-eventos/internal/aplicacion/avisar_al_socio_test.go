package aplicacion_test

import (
	"context"
	"sync"
	"testing"

	"reservas/internal/adaptadores/eventos"
	"reservas/internal/adaptadores/memoria"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

type buzonDePrueba struct {
	mu       sync.Mutex
	enviados []string
}

func (b *buzonDePrueba) Enviar(_ context.Context, correo, asunto, _ string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enviados = append(b.enviados, correo+" | "+asunto)
	return nil
}

func TestUnEventoDuplicadoSoloAvisaUnaVez(t *testing.T) {
	socios, _, _ := memoria.DatosDeEjemplo()
	buzon := &buzonDePrueba{}
	avisar := aplicacion.NuevoAvisarAlSocio(socios, buzon)
	evento := dominio.ReservaConfirmada{ReservaID: "r-1", SocioID: "s-1", PistaID: "p-1", Inicio: a(2, 10)}

	for range 2 { // la entrega es "al menos una vez": puede llegar repetido
		if err := avisar.Manejar(context.Background(), evento); err != nil {
			t.Fatal(err)
		}
	}
	if len(buzon.enviados) != 1 {
		t.Fatalf("esperaba un aviso, hubo %d: %v", len(buzon.enviados), buzon.enviados)
	}
}

// De punta a punta: reservar y cancelar publican en el bus, y el bus
// lleva los eventos a AvisarAlSocio sin que los casos de uso lo conozcan.
func TestReservarYCancelarAvisanAlSocio(t *testing.T) {
	socios, pistas, reservas := memoria.DatosDeEjemplo()
	buzon := &buzonDePrueba{}
	bus := eventos.NuevoBus(3, 0)
	bus.Suscribir("avisar-al-socio", aplicacion.NuevoAvisarAlSocio(socios, buzon).Manejar)

	relojFijo := reloj.Fijo{Momento: ahora}
	reservar := aplicacion.NuevoReservarPista(reservas, socios, pistas, relojFijo, bus)
	cancelar := aplicacion.NuevoCancelarReserva(reservas, relojFijo, bus)

	ctx := context.Background()
	dto, err := reservar.Ejecutar(ctx, aplicacion.SolicitudReserva{SocioID: "s-1", PistaID: "p-1", Inicio: a(2, 10)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cancelar.Ejecutar(ctx, aplicacion.SolicitudCancelacion{ReservaID: dto.ID, SocioID: "s-1"}); err != nil {
		t.Fatal(err)
	}
	bus.Cerrar() // espera a que el suscriptor termine

	esperado := []string{"ana@example.com | Reserva confirmada", "ana@example.com | Reserva cancelada"}
	if len(buzon.enviados) != 2 || buzon.enviados[0] != esperado[0] || buzon.enviados[1] != esperado[1] {
		t.Fatalf("avisos inesperados: %v", buzon.enviados)
	}
}

func TestReservarPublicaLaConfirmacion(t *testing.T) {
	socios, pistas, reservas := memoria.DatosDeEjemplo()
	espia := &espiaEventos{}
	reservar := aplicacion.NuevoReservarPista(reservas, socios, pistas, reloj.Fijo{Momento: ahora}, espia)
	dto, err := reservar.Ejecutar(context.Background(), aplicacion.SolicitudReserva{SocioID: "s-1", PistaID: "p-1", Inicio: a(2, 10)})
	if err != nil {
		t.Fatal(err)
	}
	if len(espia.publicados) != 1 || espia.publicados[0].Clave() != "reserva.confirmada:"+dto.ID {
		t.Fatalf("esperaba un ReservaConfirmada, salió %v", espia.publicados)
	}
}
