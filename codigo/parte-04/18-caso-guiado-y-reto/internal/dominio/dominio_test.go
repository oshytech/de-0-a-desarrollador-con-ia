package dominio

import (
	"errors"
	"testing"
	"time"
)

func fecha(hora, minuto int) time.Time {
	return time.Date(2026, 12, 1, hora, minuto, 0, 0, time.UTC)
}

func TestNuevaFranjaAplicaElHorarioDelClub(t *testing.T) {
	casos := []struct {
		nombre string
		inicio time.Time
		valida bool
	}{
		{"primera hora", fecha(8, 0), true},
		{"última hora", fecha(21, 0), true},
		{"antes de abrir", fecha(7, 0), false},
		{"acabaría después de cerrar", fecha(22, 0), false},
		{"no empieza en punto", fecha(10, 30), false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := NuevaFranja(c.inicio)
			if c.valida && err != nil {
				t.Fatalf("esperaba una franja válida y salió %v", err)
			}
			if !c.valida && !errors.Is(err, ErrFueraDeHorario) {
				t.Fatalf("esperaba ErrFueraDeHorario y salió %v", err)
			}
		})
	}
}

func TestFranjasSeguidasNoSeSolapan(t *testing.T) {
	diez, _ := NuevaFranja(fecha(10, 0))
	once, _ := NuevaFranja(fecha(11, 0))
	if diez.SeSolapaCon(once) {
		t.Fatal("10:00-11:00 y 11:00-12:00 no deberían solaparse")
	}
	if !diez.SeSolapaCon(diez) {
		t.Fatal("una franja se solapa consigo misma")
	}
}

func TestSoloReservanLosSociosActivos(t *testing.T) {
	baja := Socio{ID: "s-2", Nombre: "Luis", Activo: false}
	if !errors.Is(baja.PuedeReservar(), ErrSocioInactivo) {
		t.Fatal("un socio inactivo no debería poder reservar")
	}
}
