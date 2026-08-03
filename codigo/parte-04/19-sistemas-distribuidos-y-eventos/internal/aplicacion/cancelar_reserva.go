package aplicacion

import (
	"context"
	"fmt"

	"reservas/internal/dominio"
)

type SolicitudCancelacion struct {
	ReservaID string
	SocioID   string
}

type CancelarReserva struct {
	reservas RepositorioReservas
	reloj    Reloj
	eventos  PublicadorEventos
}

func NuevoCancelarReserva(reservas RepositorioReservas, reloj Reloj, eventos PublicadorEventos) *CancelarReserva {
	return &CancelarReserva{reservas: reservas, reloj: reloj, eventos: eventos}
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
	if err := uc.eventos.Publicar(ctx, dominio.CancelacionDe(cancelada)); err != nil {
		return ReservaDTO{}, fmt.Errorf("reserva %s cancelada, pero sin publicar: %w", cancelada.ID(), err)
	}
	return aDTO(cancelada), nil
}
