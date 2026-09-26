package aplicacion

import (
	"errors"
	"testing"
	"time"

	"reservas/internal/dominio"
)

type repoMemoria map[string]dominio.Reserva

func (m repoMemoria) Buscar(id string) (dominio.Reserva, error) {
	r, ok := m[id]
	if !ok {
		return dominio.Reserva{}, errors.New("no existe")
	}
	return r, nil
}

func (m repoMemoria) Guardar(r dominio.Reserva) error {
	m[r.ID] = r
	return nil
}

type relojFijo time.Time

func (r relojFijo) Ahora() time.Time { return time.Time(r) }

var ahora = time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)

func nuevo() (CancelarReserva, repoMemoria) {
	inicio := ahora.Add(24 * time.Hour)
	repo := repoMemoria{"r1": {ID: "r1", SocioID: "lucia", Inicio: inicio, Fin: inicio.Add(time.Hour), Estado: dominio.Confirmada}}
	return CancelarReserva{Reservas: repo, Reloj: relojFijo(ahora)}, repo
}

func TestCancelarSuReserva(t *testing.T) {
	caso, repo := nuevo()
	if err := caso.Ejecutar("lucia", "r1"); err != nil {
		t.Fatal(err)
	}
	if repo["r1"].Estado != dominio.Cancelada {
		t.Errorf("estado %q, quería cancelada", repo["r1"].Estado)
	}
}

// Reproduce el hallazgo de la revisión: otro socio cancelaba la reserva.
func TestNoSePuedeCancelarUnaReservaAjena(t *testing.T) {
	caso, repo := nuevo()
	err := caso.Ejecutar("marcos", "r1")
	if !errors.Is(err, ErrReservaAjena) {
		t.Errorf("quería ErrReservaAjena y obtuve %v", err)
	}
	if repo["r1"].Estado != dominio.Confirmada {
		t.Errorf("la reserva de otro socio ha quedado %q", repo["r1"].Estado)
	}
}
