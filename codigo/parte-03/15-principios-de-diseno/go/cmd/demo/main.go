// Raíz de composición: el único sitio que elige las piezas concretas y las conecta.
package main

import (
	"fmt"
	"time"

	"diseno/reservas"
)

func main() {
	socios := reservas.SociosEnMemoria{1: {ID: 1, Nombre: "Lucía", Activo: true}}
	reloj := reservas.RelojFijo{Hora: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)}
	reservar := reservas.NuevoReservarPista(&reservas.ReservasEnMemoria{}, socios, reloj)

	for _, hora := range []int{19, 19, 20} {
		r, err := reservar.Ejecutar(1, 1, time.Date(2026, 3, 11, hora, 0, 0, 0, time.UTC))
		if err != nil {
			fmt.Printf("%d:00 rechazada: %v\n", hora, err)
			continue
		}
		fmt.Printf("%d:00 confirmada (reserva %d)\n", hora, r.ID)
	}
}
