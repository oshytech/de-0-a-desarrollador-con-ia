package memoria

import (
	"context"
	"time"

	"reservas/internal/dominio"
)

// DatosDeEjemplo prepara un club pequeño para las demos: dos socios
// y un par de reservas para mañana a partir de la fecha que se le pase.
func DatosDeEjemplo(hoy time.Time) (*RepositorioSocios, *RepositorioReservas) {
	socios := NuevoRepositorioSocios(
		dominio.Socio{ID: "s-1", Nombre: "Ana", Correo: "ana@example.com", Activo: true},
		dominio.Socio{ID: "s-2", Nombre: "Luis", Correo: "luis@example.com", Activo: false},
	)
	manana := time.Date(hoy.Year(), hoy.Month(), hoy.Day()+1, 0, 0, 0, 0, hoy.Location())
	reservas := NuevoRepositorioReservas()
	ctx := context.Background()
	reservas.Guardar(ctx, dominio.ReconstruirReserva("r-1", "s-1", "p-1", manana.Add(18*time.Hour), dominio.Confirmada))
	reservas.Guardar(ctx, dominio.ReconstruirReserva("r-2", "s-1", "p-2", manana.Add(10*time.Hour), dominio.Confirmada))
	return socios, reservas
}
