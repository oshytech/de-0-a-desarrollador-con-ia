package dominio

import (
	"errors"
	"testing"
	"time"
)

var (
	ahora  = time.Date(2026, 12, 1, 9, 30, 0, 0, time.UTC)
	ana    = Socio{ID: "s-1", Nombre: "Ana", Activo: true}
	pista1 = Pista{ID: "p-1", Nombre: "Pista 1", Deporte: Tenis}
)

func franjaDel(t *testing.T, dia, hora int) Franja {
	t.Helper()
	f, err := NuevaFranja(time.Date(2026, 12, dia, hora, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func confirmada(id, socio, pista string, f Franja) Reserva {
	return ReconstruirReserva(id, socio, pista, f.Inicio(), Confirmada)
}

func TestNuevaReserva(t *testing.T) {
	diez := franjaDel(t, 2, 10)
	casos := []struct {
		nombre    string
		socio     Socio
		franja    Franja
		ocupacion Ocupacion
		esperado  error
	}{
		{"pista libre", ana, diez, Ocupacion{}, nil},
		{"socio de baja", Socio{ID: "s-2", Activo: false}, diez, Ocupacion{}, ErrSocioInactivo},
		{"socio de baja y pista ocupada", Socio{ID: "s-2", Activo: false}, diez,
			Ocupacion{DeLaPista: []Reserva{confirmada("r-1", "s-9", "p-1", diez)}}, ErrSocioInactivo},
		{"en el pasado", ana, franjaDel(t, 1, 9), Ocupacion{}, ErrFueraDePlazo},
		{"a más de 14 días", ana, franjaDel(t, 16, 10), Ocupacion{}, ErrFueraDePlazo},
		{"dentro de los 14 días", ana, franjaDel(t, 15, 9), Ocupacion{}, nil},
		{"pista ocupada", ana, diez, Ocupacion{DeLaPista: []Reserva{confirmada("r-1", "s-9", "p-1", diez)}}, ErrPistaOcupada},
		{"ocupada por una cancelada", ana, diez, Ocupacion{DeLaPista: []Reserva{
			ReconstruirReserva("r-1", "s-9", "p-1", diez.Inicio(), Cancelada)}}, nil},
		{"ya tiene dos futuras", ana, diez, Ocupacion{FuturasDelSocio: []Reserva{
			confirmada("r-1", "s-1", "p-2", franjaDel(t, 3, 10)),
			confirmada("r-2", "s-1", "p-2", franjaDel(t, 4, 10))}}, ErrLimiteDeReservas},
		{"las pasadas no cuentan", ana, diez, Ocupacion{FuturasDelSocio: []Reserva{
			confirmada("r-1", "s-1", "p-2", franjaDel(t, 1, 8)),
			confirmada("r-2", "s-1", "p-2", franjaDel(t, 4, 10))}}, nil},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			reserva, err := NuevaReserva("r-nueva", c.socio, pista1, c.franja.Inicio(), ahora, c.ocupacion)
			if !errors.Is(err, c.esperado) {
				t.Fatalf("esperaba %v, salió %v", c.esperado, err)
			}
			if err == nil && !reserva.EstaConfirmada() {
				t.Fatal("una reserva nueva nace confirmada")
			}
		})
	}
}

func TestCancelar(t *testing.T) {
	reserva := confirmada("r-1", "s-1", "p-1", franjaDel(t, 1, 12)) // a las 12:00, son las 9:30
	casos := []struct {
		nombre   string
		socio    string
		momento  time.Time
		esperado error
	}{
		{"con tiempo de sobra", "s-1", ahora, nil},
		{"justo 2 horas antes", "s-1", time.Date(2026, 12, 1, 10, 0, 0, 0, time.UTC), nil},
		{"a menos de 2 horas", "s-1", time.Date(2026, 12, 1, 10, 1, 0, 0, time.UTC), ErrCancelacionFueraDePlazo},
		{"de otro socio", "s-2", ahora, ErrReservaDeOtroSocio},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cancelada, err := reserva.Cancelar(c.socio, c.momento)
			if !errors.Is(err, c.esperado) {
				t.Fatalf("esperaba %v, salió %v", c.esperado, err)
			}
			if err == nil && cancelada.Estado() != Cancelada {
				t.Fatal("la reserva debería quedar cancelada")
			}
		})
	}
	if !reserva.EstaConfirmada() {
		t.Fatal("Cancelar no debe tocar la reserva original")
	}
	cancelada, _ := reserva.Cancelar("s-1", ahora)
	if _, err := cancelada.Cancelar("s-1", ahora); !errors.Is(err, ErrReservaYaCancelada) {
		t.Fatalf("cancelar dos veces debería fallar, salió %v", err)
	}
}
