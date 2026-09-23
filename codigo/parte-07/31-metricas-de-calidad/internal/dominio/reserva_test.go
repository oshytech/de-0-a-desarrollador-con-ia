package dominio

import (
	"errors"
	"testing"
	"time"
)

// Lunes 2 de marzo de 2026 a las 10:00. Todos los casos parten de aquí.
var ahora = time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)

func a(dias, hora, minuto int) time.Time {
	return time.Date(2026, 3, 2+dias, hora, minuto, 0, 0, time.UTC)
}

func confirmada(inicio time.Time) Reserva {
	return Reserva{Inicio: inicio, Fin: inicio.Add(Duracion), Estado: Confirmada}
}

func TestPuedeReservar(t *testing.T) {
	activo := Socio{ID: "s1", Nombre: "Lucía", Activo: true}
	casos := []struct {
		nombre  string
		socio   Socio
		inicio  time.Time
		dePista []Reserva
		deSocio []Reserva
		quiero  error
	}{
		{"reserva válida", activo, a(1, 18, 0), nil, nil, nil},
		{"socio inactivo", Socio{ID: "s2"}, a(1, 18, 0), nil, nil, ErrSocioInactivo},
		{"no empieza en punto", activo, a(1, 18, 30), nil, nil, ErrHorarioNoValido},
		{"antes de las 08:00", activo, a(1, 7, 0), nil, nil, ErrHorarioNoValido},
		{"la última es a las 21:00", activo, a(1, 21, 0), nil, nil, nil},
		{"a las 22:00 ya no", activo, a(1, 22, 0), nil, nil, ErrHorarioNoValido},
		{"en el pasado", activo, a(-1, 18, 0), nil, nil, ErrFueraDePlazo},
		{"justo 14 días después", activo, a(14, 10, 0), nil, nil, nil},
		{"más de 14 días", activo, a(14, 11, 0), nil, nil, ErrFueraDePlazo},
		{"pista ocupada", activo, a(1, 18, 0), []Reserva{confirmada(a(1, 18, 0))}, nil, ErrPistaOcupada},
		{"la reserva anterior acaba justo al empezar", activo, a(1, 18, 0), []Reserva{confirmada(a(1, 17, 0))}, nil, nil},
		{"una cancelada no ocupa", activo, a(1, 18, 0), []Reserva{{Inicio: a(1, 18, 0), Fin: a(1, 19, 0), Estado: Cancelada}}, nil, nil},
		{"ya tiene dos futuras", activo, a(1, 18, 0), nil, []Reserva{confirmada(a(2, 9, 0)), confirmada(a(3, 9, 0))}, ErrLimiteReservas},
		{"las pasadas no cuentan", activo, a(1, 18, 0), nil, []Reserva{confirmada(a(-2, 9, 0)), confirmada(a(3, 9, 0))}, nil},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := PuedeReservar(c.socio, c.inicio, ahora, c.dePista, c.deSocio)
			if !errors.Is(err, c.quiero) {
				t.Errorf("quería %v y obtuve %v", c.quiero, err)
			}
		})
	}
}
