// Package aplicacion contiene los casos de uso y los puertos que necesitan.
package aplicacion

import (
	"fmt"
	"time"

	"reservas/internal/dominio"
)

type Reloj interface {
	Ahora() time.Time
}

type RepositorioSocios interface {
	Buscar(id string) (dominio.Socio, error)
}

type RepositorioReservas interface {
	DePista(pistaID string) ([]dominio.Reserva, error)
	DeSocio(socioID string) ([]dominio.Reserva, error)
	Guardar(r dominio.Reserva) error
}

type ReservarPista struct {
	Reloj    Reloj
	Socios   RepositorioSocios
	Reservas RepositorioReservas
}

func (uc ReservarPista) Ejecutar(socioID, pistaID string, inicio time.Time) (dominio.Reserva, error) {
	ahora := uc.Reloj.Ahora()

	socio, err := uc.Socios.Buscar(socioID)
	if err != nil {
		return dominio.Reserva{}, err
	}
	if !socio.Activo {
		return dominio.Reserva{}, dominio.ErrSocioInactivo
	}
	if err := dominio.ComprobarHorario(inicio, ahora); err != nil {
		return dominio.Reserva{}, err
	}

	nueva := dominio.Reserva{
		ID:      fmt.Sprintf("%s-%s", pistaID, inicio.Format("200601021504")),
		SocioID: socioID,
		PistaID: pistaID,
		Inicio:  inicio,
		Fin:     inicio.Add(dominio.DuracionReserva),
		Estado:  dominio.Confirmada,
	}

	dePista, err := uc.Reservas.DePista(pistaID)
	if err != nil {
		return dominio.Reserva{}, err
	}
	for _, r := range dePista {
		if r.Estado == dominio.Confirmada && dominio.SeSolapan(r, nueva) {
			return dominio.Reserva{}, dominio.ErrPistaOcupada
		}
	}

	deSocio, err := uc.Reservas.DeSocio(socioID)
	if err != nil {
		return dominio.Reserva{}, err
	}
	futuras := 0
	for _, r := range deSocio {
		if r.Estado == dominio.Confirmada && r.Inicio.After(ahora) {
			futuras++
		}
	}
	if err := dominio.ComprobarLimite(futuras); err != nil {
		return dominio.Reserva{}, err
	}

	return nueva, uc.Reservas.Guardar(nueva)
}
