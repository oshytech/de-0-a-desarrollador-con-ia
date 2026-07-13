package reservas

import (
	"errors"
	"testing"
	"time"
)

var ahora = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func nuevoCaso() *ReservarPista {
	socios := SociosEnMemoria{
		1: {ID: 1, Nombre: "Lucía", Activo: true},
		2: {ID: 2, Nombre: "Marcos", Activo: true},
		3: {ID: 3, Nombre: "Inés", Activo: false},
	}
	return NuevoReservarPista(&ReservasEnMemoria{}, socios, RelojFijo{Hora: ahora})
}

func a(dia, hora int) time.Time { return time.Date(2026, 3, dia, hora, 0, 0, 0, time.UTC) }

func TestReservaValida(t *testing.T) {
	r, err := nuevoCaso().Ejecutar(1, 1, a(11, 19))
	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if r.ID == 0 || r.Estado != "confirmada" {
		t.Errorf("reserva mal guardada: %+v", r)
	}
}

func TestReglas(t *testing.T) {
	casos := []struct {
		nombre  string
		previas [][3]int // socio, día, hora de reservas que ya existen en la pista 1
		socio   int
		inicio  time.Time
		quiero  error
	}{
		{"socio inactivo", nil, 3, a(11, 9), ErrSocioInactivo},
		{"socio que no existe", nil, 9, a(11, 9), ErrSocioNoExiste},
		{"antes de abrir", nil, 1, a(11, 7), ErrFueraDeHorario},
		{"empieza a las 22", nil, 1, a(11, 22), ErrFueraDeHorario},
		{"no en punto", nil, 1, a(11, 9).Add(30 * time.Minute), ErrFueraDeHorario},
		{"en el pasado", nil, 1, a(10, 9), ErrFueraDePlazo},
		{"más de 14 días", nil, 1, a(25, 9), ErrFueraDePlazo},
		{"pista ocupada", [][3]int{{2, 11, 19}}, 1, a(11, 19), ErrPistaOcupada},
		{"tercera reserva", [][3]int{{1, 11, 9}, {1, 12, 9}}, 1, a(13, 9), ErrLimiteReservas},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			uc := nuevoCaso()
			for _, p := range c.previas {
				if _, err := uc.Ejecutar(p[0], 1, a(p[1], p[2])); err != nil {
					t.Fatalf("preparando el caso: %v", err)
				}
			}
			_, err := uc.Ejecutar(c.socio, 1, c.inicio)
			if !errors.Is(err, c.quiero) {
				t.Errorf("quería %v y obtuve %v", c.quiero, err)
			}
		})
	}
}
