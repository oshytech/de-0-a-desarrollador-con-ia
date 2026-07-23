// Otro adaptador de entrada para el mismo caso de uso: la terminal.
// Uso: go run ./cmd/consola s-1
package main

import (
	"context"
	"fmt"
	"os"
	"time"
	_ "time/tzdata"

	"reservas/internal/adaptadores/memoria"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "uso: consola <id-de-socio>")
		os.Exit(2)
	}
	zona, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	relojDelClub := reloj.Sistema{}
	socios, reservas := memoria.DatosDeEjemplo(relojDelClub.Ahora().In(zona))
	listar := aplicacion.NuevoListarReservasDeSocio(socios, reservas, relojDelClub)

	proximas, err := listar.Ejecutar(context.Background(), os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	for _, r := range proximas {
		fmt.Printf("%s  pista %s  %s  (%s)\n", r.ID, r.PistaID, r.Inicio.In(zona).Format("02/01 15:04"), r.Estado)
	}
}
