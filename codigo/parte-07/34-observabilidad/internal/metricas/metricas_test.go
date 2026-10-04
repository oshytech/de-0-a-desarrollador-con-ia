package metricas

import (
	"testing"
	"time"
)

func TestPercentil(t *testing.T) {
	ms := func(ns ...int) []time.Duration {
		ds := make([]time.Duration, len(ns))
		for i, n := range ns {
			ds[i] = time.Duration(n) * time.Millisecond
		}
		return ds
	}
	// Diez peticiones: nueve rápidas y una lenta.
	diez := ms(12, 10, 11, 9, 13, 10, 12, 11, 10, 900)
	casos := []struct {
		nombre string
		ds     []time.Duration
		p      float64
		quiero time.Duration
	}{
		{"sin datos", nil, 95, 0},
		{"mediana", diez, 50, 11 * time.Millisecond},
		{"p90", diez, 90, 13 * time.Millisecond},
		{"p99 ve la lenta", diez, 99, 900 * time.Millisecond},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := Percentil(c.ds, c.p); got != c.quiero {
				t.Fatalf("Percentil(%v) = %s, quería %s", c.p, got, c.quiero)
			}
		})
	}
}
