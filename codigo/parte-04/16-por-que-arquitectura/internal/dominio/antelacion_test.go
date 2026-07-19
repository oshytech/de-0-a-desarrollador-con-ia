package dominio

import (
	"errors"
	"testing"
	"time"
)

func TestComprobarAntelacion(t *testing.T) {
	ahora := time.Date(2026, 12, 1, 9, 30, 0, 0, time.UTC)
	casos := []struct {
		nombre string
		inicio time.Time
		valida bool
	}{
		{"mañana", time.Date(2026, 12, 2, 10, 0, 0, 0, time.UTC), true},
		{"hace una hora", time.Date(2026, 12, 1, 8, 30, 0, 0, time.UTC), false},
		{"dentro de 14 días justos", ahora.Add(MaxAntelacion), true},
		{"un minuto más tarde", ahora.Add(MaxAntelacion + time.Minute), false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := ComprobarAntelacion(c.inicio, ahora)
			if c.valida != (err == nil) || (err != nil && !errors.Is(err, ErrFueraDePlazo)) {
				t.Fatalf("resultado inesperado: %v", err)
			}
		})
	}
}
