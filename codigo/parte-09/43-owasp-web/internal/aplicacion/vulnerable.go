package aplicacion

import "context"

// CancelarReservaVulnerable es la versión con control de acceso roto (A01).
// Se deja en el repositorio solo para que los tests demuestren el fallo.
// Comprueba que la sesión es válida (eso lo hace el middleware) y las reglas
// de negocio, pero nunca relaciona al socio con la reserva.
type CancelarReservaVulnerable struct {
	Reservas RepositorioReservas
	Reloj    Reloj
}

func (c CancelarReservaVulnerable) Ejecutar(ctx context.Context, _ int64, reservaID int64) error {
	r, err := c.Reservas.Buscar(ctx, reservaID)
	if err != nil {
		return err
	}
	if err := r.Cancelar(c.Reloj.Ahora()); err != nil {
		return err
	}
	return c.Reservas.Guardar(ctx, r)
}
