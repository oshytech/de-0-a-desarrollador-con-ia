//go:build sincomprobar

package dominio

import "testing"

// Ejemplo de lo que NO hay que hacer: recorre todas las ramas y no comprueba nada.
// Se compila solo con: go test -tags sincomprobar -run TestSinComprobar -cover ./internal/dominio/
func TestSinComprobar(t *testing.T) {
	activo := Socio{Activo: true}
	PuedeReservar(Socio{}, a(1, 18, 0), ahora, nil, nil)
	PuedeReservar(activo, a(1, 18, 30), ahora, nil, nil)
	PuedeReservar(activo, a(1, 7, 0), ahora, nil, nil)
	PuedeReservar(activo, a(-1, 18, 0), ahora, nil, nil)
	PuedeReservar(activo, a(1, 18, 0), ahora, []Reserva{confirmada(a(1, 18, 0))}, nil)
	PuedeReservar(activo, a(1, 18, 0), ahora, nil, []Reserva{confirmada(a(2, 9, 0)), confirmada(a(3, 9, 0))})
	PuedeReservar(activo, a(1, 18, 0), ahora, nil, nil)
	PuedeCancelar(confirmada(a(0, 11, 0)), ahora)
	PuedeCancelar(confirmada(a(1, 11, 0)), ahora)
}
