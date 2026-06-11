// Este programa no compila a propósito: es el ejemplo del capítulo 7.
// Ejecútalo con "go run ." para ver el error del compilador.
package main

import "fmt"

func totalConIVA(precio float64) float64 {
	return precio * 1.21
}

func main() {
	fmt.Println(totalConIVA(100))
	fmt.Println(totalConIVA("100"))
}
