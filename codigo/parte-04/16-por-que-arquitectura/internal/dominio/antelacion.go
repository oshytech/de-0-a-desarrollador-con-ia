// Package dominio es la primera extracción: la regla de antelación sola,
// sin HTTP, sin JSON y sin preguntar la hora al sistema.
package dominio

import (
	"errors"
	"time"
)

var ErrFueraDePlazo = errors.New("no se reserva en el pasado ni con más de 14 días de antelación")

const MaxAntelacion = 14 * 24 * time.Hour

// ComprobarAntelacion aplica la regla 6. La hora actual llega como
// parámetro, así que el resultado solo depende de lo que se le pasa.
func ComprobarAntelacion(inicio, ahora time.Time) error {
	if !inicio.After(ahora) || inicio.Sub(ahora) > MaxAntelacion {
		return ErrFueraDePlazo
	}
	return nil
}
