package reservas

import "time"

// ReservasEnMemoria cumple RepositorioReservas sin declararlo en ninguna parte.
type ReservasEnMemoria struct {
	datos []Reserva
}

func (m *ReservasEnMemoria) DePista(pistaID int, dia time.Time) ([]Reserva, error) {
	var resultado []Reserva
	for _, r := range m.datos {
		if r.PistaID == pistaID && r.Inicio.YearDay() == dia.YearDay() && r.Inicio.Year() == dia.Year() {
			resultado = append(resultado, r)
		}
	}
	return resultado, nil
}

func (m *ReservasEnMemoria) FuturasDeSocio(socioID int, desde time.Time) ([]Reserva, error) {
	var resultado []Reserva
	for _, r := range m.datos {
		if r.SocioID == socioID && r.Estado == "confirmada" && r.Inicio.After(desde) {
			resultado = append(resultado, r)
		}
	}
	return resultado, nil
}

func (m *ReservasEnMemoria) Guardar(r Reserva) (Reserva, error) {
	r.ID = len(m.datos) + 1
	m.datos = append(m.datos, r)
	return r, nil
}

type SociosEnMemoria map[int]Socio

func (s SociosEnMemoria) Buscar(id int) (Socio, bool, error) {
	socio, ok := s[id]
	return socio, ok, nil
}

// RelojDelSistema da la hora real. RelojFijo, la que le digas: sirve para los tests.
type RelojDelSistema struct{}

func (RelojDelSistema) Ahora() time.Time { return time.Now() }

type RelojFijo struct{ Hora time.Time }

func (r RelojFijo) Ahora() time.Time { return r.Hora }

// Comprobaciones en tiempo de compilación: si dejan de cumplir la interfaz, no compila.
var (
	_ RepositorioReservas = (*ReservasEnMemoria)(nil)
	_ RepositorioSocios   = SociosEnMemoria(nil)
	_ Reloj               = RelojFijo{}
)
