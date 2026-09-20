package tarifas

import (
	"errors"
	"testing"
	"time"
)

func TestPrecio(t *testing.T) {
	martes10 := time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC)
	martes19 := time.Date(2026, 3, 3, 19, 0, 0, 0, time.UTC)
	sabado19 := time.Date(2026, 3, 7, 19, 0, 0, 0, time.UTC)

	casos := []struct {
		nombre  string
		deporte Deporte
		inicio  time.Time
		extras  Extras
		quiero  Centimos
		err     error
	}{
		{"tenis entre semana por la mañana", Tenis, martes10, Extras{}, 1200, nil},
		{"pádel en hora punta", Padel, martes19, Extras{}, 2000, nil},
		{"pádel en sábado por la tarde con luz", Padel, sabado19, Extras{ConLuz: true}, 2500, nil},
		{"el bono descuenta un 20 % del total", Padel, sabado19, Extras{ConLuz: true, ConBono: true}, 2000, nil},
		{"deporte que el club no tiene", Deporte("squash"), martes10, Extras{}, 0, ErrDeporteDesconocido},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got, err := Precio(c.deporte, c.inicio, c.extras)
			if got != c.quiero || !errors.Is(err, c.err) {
				t.Errorf("Precio = %d, %v; quiero %d, %v", got, err, c.quiero, c.err)
			}
		})
	}
}
