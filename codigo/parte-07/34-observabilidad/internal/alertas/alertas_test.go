package alertas

import (
	"testing"
	"time"
)

func TestRevisarDuracion(t *testing.T) {
	casos := []struct {
		nombre   string
		duracion time.Duration
		avisa    bool
	}{
		{"normal", 40 * time.Minute, false},
		{"justo en el límite", 60 * time.Minute, false},
		{"se ha pasado", 61 * time.Minute, true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, avisa := RevisarDuracion("recordatorios", c.duracion, 40*time.Minute, 1.5)
			if avisa != c.avisa {
				t.Fatalf("avisa = %v, quería %v", avisa, c.avisa)
			}
		})
	}
}

func TestAgrupadorCallaLasRepetidas(t *testing.T) {
	g := NuevoAgrupador(10 * time.Minute)
	a := Alerta{Clave: "limite-api"}
	t0 := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)

	if enviar, _ := g.Decidir(a, t0); !enviar {
		t.Fatal("la primera alerta tiene que enviarse")
	}
	for i := 1; i <= 5; i++ {
		if enviar, _ := g.Decidir(a, t0.Add(time.Duration(i)*time.Minute)); enviar {
			t.Fatalf("la repetición %d no debería enviarse", i)
		}
	}
	enviar, calladas := g.Decidir(a, t0.Add(11*time.Minute))
	if !enviar || calladas != 5 {
		t.Fatalf("tras la ventana: enviar = %v, calladas = %d; quería true y 5", enviar, calladas)
	}
	if enviar, _ := g.Decidir(Alerta{Clave: "otra"}, t0.Add(2*time.Minute)); !enviar {
		t.Fatal("una clave distinta no se agrupa con la anterior")
	}
}
