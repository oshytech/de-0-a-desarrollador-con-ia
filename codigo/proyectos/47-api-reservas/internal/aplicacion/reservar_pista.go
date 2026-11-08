package aplicacion

import (
	"context"
	"fmt"
	"time"

	"reservas/internal/dominio"
)

// SolicitudReserva es la entrada del caso de uso: datos sencillos,
// sin tipos del dominio ni de HTTP.
type SolicitudReserva struct {
	SocioID string
	PistaID string
	Inicio  time.Time
}

type ReservarPista struct {
	reservas RepositorioReservas
	socios   RepositorioSocios
	pistas   RepositorioPistas
	reloj    Reloj
}

func NuevoReservarPista(reservas RepositorioReservas, socios RepositorioSocios, pistas RepositorioPistas, reloj Reloj) *ReservarPista {
	return &ReservarPista{reservas: reservas, socios: socios, pistas: pistas, reloj: reloj}
}

func (uc *ReservarPista) Ejecutar(ctx context.Context, s SolicitudReserva) (ReservaDTO, error) {
	// 1. Cargar lo que hace falta para decidir.
	socio, err := uc.socios.BuscarPorID(ctx, s.SocioID)
	if err != nil {
		return ReservaDTO{}, fmt.Errorf("socio %q: %w", s.SocioID, err)
	}
	pista, err := uc.pistas.BuscarPorID(ctx, s.PistaID)
	if err != nil {
		return ReservaDTO{}, fmt.Errorf("pista %q: %w", s.PistaID, err)
	}
	ahora := uc.reloj.Ahora()
	deLaPista, err := uc.reservas.ConfirmadasDePista(ctx, pista.ID, s.Inicio, s.Inicio.Add(dominio.Duracion))
	if err != nil {
		return ReservaDTO{}, fmt.Errorf("reservas de la pista %q: %w", pista.ID, err)
	}
	delSocio, err := uc.reservas.ConfirmadasDeSocioDesde(ctx, socio.ID, ahora)
	if err != nil {
		return ReservaDTO{}, fmt.Errorf("reservas del socio %q: %w", socio.ID, err)
	}
	// 2. Que decida el dominio.
	ocupacion := dominio.Ocupacion{DeLaPista: deLaPista, FuturasDelSocio: delSocio}
	reserva, err := dominio.NuevaReserva(uc.reservas.NuevoID(), socio, pista, s.Inicio, ahora, ocupacion)
	if err != nil {
		return ReservaDTO{}, err
	}
	// 3. Guardar el resultado.
	if err := uc.reservas.Guardar(ctx, reserva); err != nil {
		return ReservaDTO{}, fmt.Errorf("guardar reserva: %w", err)
	}
	// 4. Devolver una respuesta que no expone la entidad.
	return aDTO(reserva), nil
}
