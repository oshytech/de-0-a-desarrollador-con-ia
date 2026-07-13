package reservas

import (
	"errors"
	"time"
)

// Mal: el archivo histórico es de solo lectura, pero se le obliga a cumplir
// RepositorioReservas. Compila, y quien lo use descubrirá en ejecución que Guardar falla.
type HistoricoQueMiente struct{ ReservasEnMemoria }

func (HistoricoQueMiente) Guardar(r Reserva) (Reserva, error) {
	return Reserva{}, errors.New("el histórico es de solo lectura")
}

// Bien: una interfaz pequeña para quien solo consulta.
type LectorReservas interface {
	FuturasDeSocio(socioID int, desde time.Time) ([]Reserva, error)
}

type Historico struct{ datos []Reserva }

func (h Historico) FuturasDeSocio(socioID int, desde time.Time) ([]Reserva, error) {
	var resultado []Reserva
	for _, r := range h.datos {
		if r.SocioID == socioID && r.Inicio.After(desde) {
			resultado = append(resultado, r)
		}
	}
	return resultado, nil
}

var (
	_ RepositorioReservas = (*HistoricoQueMiente)(nil) // compila, y ese es el problema
	_ LectorReservas      = Historico{}
	_ LectorReservas      = (*ReservasEnMemoria)(nil) // el repositorio completo también sirve
)
