package aplicacion

import (
	"context"
	"fmt"
)

type SolicitudCancelacion struct {
	ReservaID string
	SocioID   string
}

type CancelarReserva struct {
	reservas RepositorioReservas
	reloj    Reloj
}

func NuevoCancelarReserva(reservas RepositorioReservas, reloj Reloj) *CancelarReserva {
	return &CancelarReserva{reservas: reservas, reloj: reloj}
}

func (uc *CancelarReserva) Ejecutar(ctx context.Context, s SolicitudCancelacion) (ReservaDTO, error) {
	reserva, err := uc.reservas.BuscarPorID(ctx, s.ReservaID)
	if err != nil {
		return ReservaDTO{}, fmt.Errorf("reserva %q: %w", s.ReservaID, err)
	}
	cancelada, err := reserva.Cancelar(s.SocioID, uc.reloj.Ahora())
	if err != nil {
		return ReservaDTO{}, err
	}
	if err := uc.reservas.Guardar(ctx, cancelada); err != nil {
		return ReservaDTO{}, fmt.Errorf("guardar reserva: %w", err)
	}
	return aDTO(cancelada), nil
}
