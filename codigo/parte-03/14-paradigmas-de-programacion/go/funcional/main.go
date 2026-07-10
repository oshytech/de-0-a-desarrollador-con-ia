// Funciones como valores en Go: filtros genéricos y closures.
package main

import (
	"fmt"
	"time"
)

type Reserva struct {
	Socio  string
	Inicio time.Time
	Estado string
}

// filtrar sirve para cualquier tipo gracias a los genéricos.
func filtrar[T any](xs []T, cumple func(T) bool) []T {
	var resultado []T
	for _, x := range xs {
		if cumple(x) {
			resultado = append(resultado, x)
		}
	}
	return resultado
}

// futurasConfirmadasDe devuelve una función que "recuerda" el socio y la hora.
func futurasConfirmadasDe(socio string, ahora time.Time) func(Reserva) bool {
	return func(r Reserva) bool {
		return r.Socio == socio && r.Estado == "confirmada" && r.Inicio.After(ahora)
	}
}

func main() {
	ahora := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	reservas := []Reserva{
		{"Lucía", time.Date(2026, 3, 9, 18, 0, 0, 0, time.UTC), "confirmada"},
		{"Lucía", time.Date(2026, 3, 11, 9, 0, 0, 0, time.UTC), "confirmada"},
		{"Marcos", time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC), "confirmada"},
		{"Lucía", time.Date(2026, 3, 12, 20, 0, 0, 0, time.UTC), "cancelada"},
	}
	fmt.Println(len(filtrar(reservas, futurasConfirmadasDe("Lucía", ahora))))
}
