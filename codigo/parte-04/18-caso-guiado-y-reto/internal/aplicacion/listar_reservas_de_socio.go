package aplicacion

import (
	"context"
	"fmt"
	"slices"
)

// ListarReservasDeSocio devuelve las reservas de un socio que todavía no han empezado.
type ListarReservasDeSocio struct {
	socios   RepositorioSocios
	reservas RepositorioReservas
	reloj    Reloj
}

func NuevoListarReservasDeSocio(socios RepositorioSocios, reservas RepositorioReservas, reloj Reloj) *ListarReservasDeSocio {
	return &ListarReservasDeSocio{socios: socios, reservas: reservas, reloj: reloj}
}

func (uc *ListarReservasDeSocio) Ejecutar(ctx context.Context, socioID string) ([]ReservaDTO, error) {
	if _, err := uc.socios.BuscarPorID(ctx, socioID); err != nil {
		return nil, fmt.Errorf("socio %q: %w", socioID, err)
	}
	todas, err := uc.reservas.DeSocio(ctx, socioID)
	if err != nil {
		return nil, fmt.Errorf("reservas del socio %q: %w", socioID, err)
	}
	ahora := uc.reloj.Ahora()
	proximas := []ReservaDTO{}
	for _, r := range todas {
		if r.Franja().Inicio().After(ahora) {
			proximas = append(proximas, aDTO(r))
		}
	}
	slices.SortFunc(proximas, func(a, b ReservaDTO) int { return a.Inicio.Compare(b.Inicio) })
	return proximas, nil
}
