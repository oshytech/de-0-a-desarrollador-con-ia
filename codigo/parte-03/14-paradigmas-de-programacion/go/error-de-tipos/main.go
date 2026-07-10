//go:build ignore

// Este programa no compila a propósito: Go comprueba los tipos antes de ejecutar nada.
package main

import "fmt"

func finDeReserva(horaInicio int) int {
	return horaInicio + 1
}

func main() {
	fmt.Println(finDeReserva(9))
	fmt.Println(finDeReserva("9"))
}
