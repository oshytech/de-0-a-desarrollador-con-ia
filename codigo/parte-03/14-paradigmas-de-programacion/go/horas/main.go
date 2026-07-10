// Horas libres de una pista, en Go.
package main

import "fmt"

const (
	apertura     = 8  // primera hora en la que puede empezar una reserva
	ultimoInicio = 21 // la última reserva empieza a las 21:00
)

func horasLibres(ocupadas map[int]bool) []int {
	var libres []int
	for hora := apertura; hora <= ultimoInicio; hora++ {
		if !ocupadas[hora] {
			libres = append(libres, hora)
		}
	}
	return libres
}

func main() {
	ocupadas := map[int]bool{9: true, 10: true, 18: true, 19: true, 20: true}
	fmt.Println(horasLibres(ocupadas))
}
