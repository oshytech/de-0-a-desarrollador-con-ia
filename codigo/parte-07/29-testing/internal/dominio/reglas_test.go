package dominio

import (
	"errors"
	"testing"
	"time"
)

// Un lunes cualquiera a las 09:30, fijo para que los tests sean deterministas.
var ahora = time.Date(2026, 3, 2, 9, 30, 0, 0, time.UTC)

func en(dias, hora, minuto int) time.Time {
	return time.Date(2026, 3, 2+dias, hora, minuto, 0, 0, time.UTC)
}

func TestTablaComprobarHorario(t *testing.T) {
	casos := []struct {
		nombre string
		inicio time.Time
		quiero error
	}{
		{"primera hora del día", en(1, 8, 0), nil},
		{"última hora del día", en(1, 21, 0), nil},
		{"antes de abrir", en(1, 7, 0), ErrHorarioNoValido},
		{"después de la última", en(1, 22, 0), ErrHorarioNoValido},
		{"no empieza en punto", en(1, 10, 30), ErrHorarioNoValido},
		{"hoy, más tarde", en(0, 11, 0), nil},
		{"hoy, ya pasada", en(0, 9, 0), ErrFueraDePlazo},
		{"justo dentro de 14 días", en(14, 9, 0), nil},
		{"pasados los 14 días", en(14, 10, 0), ErrFueraDePlazo},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if err := ComprobarHorario(c.inicio, ahora); !errors.Is(err, c.quiero) {
				t.Errorf("ComprobarHorario(%v) = %v, quiero %v", c.inicio, err, c.quiero)
			}
		})
	}
}

func TestTablaSeSolapan(t *testing.T) {
	a := Reserva{Inicio: en(1, 10, 0), Fin: en(1, 11, 0)}
	casos := []struct {
		nombre string
		b      Reserva
		quiero bool
	}{
		{"misma hora", Reserva{Inicio: en(1, 10, 0), Fin: en(1, 11, 0)}, true},
		{"empieza cuando acaba la otra", Reserva{Inicio: en(1, 11, 0), Fin: en(1, 12, 0)}, false},
		{"acaba cuando empieza la otra", Reserva{Inicio: en(1, 9, 0), Fin: en(1, 10, 0)}, false},
		{"otro día", Reserva{Inicio: en(2, 10, 0), Fin: en(2, 11, 0)}, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := SeSolapan(a, c.b); got != c.quiero {
				t.Errorf("SeSolapan = %v, quiero %v", got, c.quiero)
			}
		})
	}
}

func TestTablaComprobarLimite(t *testing.T) {
	casos := []struct {
		confirmadas int
		quiero      error
	}{
		{0, nil},
		{1, nil},
		{2, ErrLimiteReservas},
	}
	for _, c := range casos {
		if err := ComprobarLimite(c.confirmadas); !errors.Is(err, c.quiero) {
			t.Errorf("ComprobarLimite(%d) = %v, quiero %v", c.confirmadas, err, c.quiero)
		}
	}
}

func TestTablaComprobarCancelacion(t *testing.T) {
	casos := []struct {
		nombre string
		inicio time.Time
		quiero error
	}{
		{"con mucho margen", en(0, 15, 0), nil},
		{"exactamente 2 horas antes", en(0, 11, 30), nil},
		{"un minuto tarde", en(0, 11, 29), ErrCancelacionTardia},
		{"ya empezada", en(0, 9, 0), ErrCancelacionTardia},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r := Reserva{Inicio: c.inicio, Fin: c.inicio.Add(DuracionReserva), Estado: Confirmada}
			if err := ComprobarCancelacion(r, ahora); !errors.Is(err, c.quiero) {
				t.Errorf("ComprobarCancelacion = %v, quiero %v", err, c.quiero)
			}
		})
	}
}

// Añadido después de ver sobrevivir al mutante del plazo: con "ahora" a las 09:30
// ninguna reserva en punto cae justo a 14 días, así que hace falta otro reloj.
func TestFronteraPlazo(t *testing.T) {
	ahoraEnPunto := time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC)
	justo14Dias := ahoraEnPunto.Add(AntelacionMaxima)
	if err := ComprobarHorario(justo14Dias, ahoraEnPunto); err != nil {
		t.Errorf("a 14 días justos debería valer, da %v", err)
	}
}
