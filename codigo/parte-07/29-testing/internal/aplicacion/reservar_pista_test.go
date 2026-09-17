package aplicacion

import (
	"errors"
	"testing"
	"time"

	"reservas/internal/dominio"
)

// Dobles de prueba: un reloj que siempre da la misma hora y repositorios en memoria.

type relojFijo struct{ t time.Time }

func (r relojFijo) Ahora() time.Time { return r.t }

type sociosEnMemoria map[string]dominio.Socio

func (s sociosEnMemoria) Buscar(id string) (dominio.Socio, error) {
	socio, ok := s[id]
	if !ok {
		return dominio.Socio{}, errors.New("socio no encontrado")
	}
	return socio, nil
}

type reservasEnMemoria struct{ guardadas []dominio.Reserva }

func (m *reservasEnMemoria) DePista(id string) ([]dominio.Reserva, error) {
	var res []dominio.Reserva
	for _, r := range m.guardadas {
		if r.PistaID == id {
			res = append(res, r)
		}
	}
	return res, nil
}

func (m *reservasEnMemoria) DeSocio(id string) ([]dominio.Reserva, error) {
	var res []dominio.Reserva
	for _, r := range m.guardadas {
		if r.SocioID == id {
			res = append(res, r)
		}
	}
	return res, nil
}

func (m *reservasEnMemoria) Guardar(r dominio.Reserva) error {
	m.guardadas = append(m.guardadas, r)
	return nil
}

var lunes = time.Date(2026, 3, 2, 9, 30, 0, 0, time.UTC)

func a(dias, hora int) time.Time {
	return time.Date(2026, 3, 2+dias, hora, 0, 0, 0, time.UTC)
}

func reserva(socio, pista string, inicio time.Time, estado dominio.EstadoReserva) dominio.Reserva {
	return dominio.Reserva{SocioID: socio, PistaID: pista, Inicio: inicio,
		Fin: inicio.Add(time.Hour), Estado: estado}
}

func TestReservarPista(t *testing.T) {
	socios := sociosEnMemoria{
		"ana":  {ID: "ana", Nombre: "Ana", Activo: true},
		"luis": {ID: "luis", Nombre: "Luis", Activo: false},
	}
	casos := []struct {
		nombre  string
		previas []dominio.Reserva
		socio   string
		inicio  time.Time
		quiero  error
	}{
		{"pista libre", nil, "ana", a(1, 10), nil},
		{"socio inactivo", nil, "luis", a(1, 10), dominio.ErrSocioInactivo},
		{"fuera de horario", nil, "ana", a(1, 22), dominio.ErrHorarioNoValido},
		{"en el pasado", nil, "ana", a(-1, 10), dominio.ErrFueraDePlazo},
		{"pista ocupada",
			[]dominio.Reserva{reserva("otro", "p1", a(1, 10), dominio.Confirmada)},
			"ana", a(1, 10), dominio.ErrPistaOcupada},
		{"la reserva cancelada no ocupa",
			[]dominio.Reserva{reserva("otro", "p1", a(1, 10), dominio.Cancelada)},
			"ana", a(1, 10), nil},
		{"ya tiene dos futuras",
			[]dominio.Reserva{
				reserva("ana", "p2", a(2, 10), dominio.Confirmada),
				reserva("ana", "p2", a(3, 10), dominio.Confirmada),
			},
			"ana", a(1, 10), dominio.ErrLimiteReservas},
		{"las pasadas no cuentan para el límite",
			[]dominio.Reserva{
				reserva("ana", "p2", a(-2, 10), dominio.Confirmada),
				reserva("ana", "p2", a(3, 10), dominio.Confirmada),
			},
			"ana", a(1, 10), nil},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := &reservasEnMemoria{guardadas: c.previas}
			uc := ReservarPista{Reloj: relojFijo{lunes}, Socios: socios, Reservas: repo}

			_, err := uc.Ejecutar(c.socio, "p1", c.inicio)

			if !errors.Is(err, c.quiero) {
				t.Fatalf("error = %v, quiero %v", err, c.quiero)
			}
			if c.quiero == nil && len(repo.guardadas) != len(c.previas)+1 {
				t.Errorf("no se ha guardado la reserva")
			}
		})
	}
}
