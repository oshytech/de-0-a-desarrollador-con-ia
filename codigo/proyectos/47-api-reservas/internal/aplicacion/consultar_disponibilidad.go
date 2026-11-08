package aplicacion

import (
	"context"
	"fmt"
	"time"

	"reservas/internal/dominio"
)

// Disponibilidad son las horas de inicio que se podrían reservar ahora
// mismo en una pista y un día concretos.
type Disponibilidad struct {
	Fecha       time.Time
	HorasLibres []time.Time
}

// ConsultarDisponibilidad carga lo necesario y deja que el dominio decida
// qué horas quedan libres (dominio.HorasLibres).
type ConsultarDisponibilidad struct {
	pistas   RepositorioPistas
	reservas RepositorioReservas
	reloj    Reloj
}

func NuevoConsultarDisponibilidad(pistas RepositorioPistas, reservas RepositorioReservas, reloj Reloj) *ConsultarDisponibilidad {
	return &ConsultarDisponibilidad{pistas: pistas, reservas: reservas, reloj: reloj}
}

// Ejecutar recibe el día en la zona horaria del club, igual que ReservarPista
// recibe el inicio: el adaptador HTTP es quien conoce la zona.
func (uc *ConsultarDisponibilidad) Ejecutar(ctx context.Context, pistaID string, dia time.Time) (Disponibilidad, error) {
	pista, err := uc.pistas.BuscarPorID(ctx, pistaID)
	if err != nil {
		return Disponibilidad{}, fmt.Errorf("pista %q: %w", pistaID, err)
	}
	medianoche := time.Date(dia.Year(), dia.Month(), dia.Day(), 0, 0, 0, 0, dia.Location())
	ocupadas, err := uc.reservas.ConfirmadasDePista(ctx, pista.ID, medianoche, medianoche.AddDate(0, 0, 1))
	if err != nil {
		return Disponibilidad{}, fmt.Errorf("reservas de la pista %q: %w", pista.ID, err)
	}
	disponibilidad := Disponibilidad{Fecha: medianoche, HorasLibres: []time.Time{}}
	for _, franja := range dominio.HorasLibres(medianoche, uc.reloj.Ahora(), ocupadas) {
		disponibilidad.HorasLibres = append(disponibilidad.HorasLibres, franja.Inicio())
	}
	return disponibilidad, nil
}
