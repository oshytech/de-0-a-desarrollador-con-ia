// Package reloj tiene los adaptadores del puerto aplicacion.Reloj.
package reloj

import "time"

// Sistema da la hora real.
type Sistema struct{}

func (Sistema) Ahora() time.Time { return time.Now() }

// Fijo siempre da la misma hora. Sirve para tests y demos reproducibles.
type Fijo struct {
	Momento time.Time
}

func (f Fijo) Ahora() time.Time { return f.Momento }
