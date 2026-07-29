package memoria

import "reservas/internal/dominio"

// DatosDeEjemplo prepara un club pequeño para las demos y los tests:
// tres socios (uno de baja), dos pistas y ninguna reserva.
func DatosDeEjemplo() (*RepositorioSocios, *RepositorioPistas, *RepositorioReservas) {
	socios := NuevoRepositorioSocios(
		dominio.Socio{ID: "s-1", Nombre: "Ana", Correo: "ana@example.com", Activo: true},
		dominio.Socio{ID: "s-2", Nombre: "Luis", Correo: "luis@example.com", Activo: false},
		dominio.Socio{ID: "s-3", Nombre: "Marta", Correo: "marta@example.com", Activo: true},
	)
	pistas := NuevoRepositorioPistas(
		dominio.Pista{ID: "p-1", Nombre: "Pista 1", Deporte: dominio.Tenis},
		dominio.Pista{ID: "p-2", Nombre: "Pista 2", Deporte: dominio.Padel},
	)
	return socios, pistas, NuevoRepositorioReservas()
}
