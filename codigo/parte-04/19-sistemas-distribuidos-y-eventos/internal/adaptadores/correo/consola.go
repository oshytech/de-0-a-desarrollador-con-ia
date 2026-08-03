// Package correo tiene los adaptadores del puerto aplicacion.Notificador.
package correo

import (
	"context"
	"fmt"
	"io"
	"sync"

	"reservas/internal/aplicacion"
)

// Consola no envía nada: escribe el mensaje para que se vea en la demo.
type Consola struct {
	mu     sync.Mutex
	salida io.Writer
}

var _ aplicacion.Notificador = (*Consola)(nil)

func NuevaConsola(salida io.Writer) *Consola { return &Consola{salida: salida} }

func (c *Consola) Enviar(_ context.Context, correo, asunto, texto string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := fmt.Fprintf(c.salida, "[correo] para %s | %s | %s\n", correo, asunto, texto)
	return err
}
