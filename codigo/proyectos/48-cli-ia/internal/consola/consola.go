// Package consola es el adaptador de entrada: lee preguntas de la terminal
// y escribe las respuestas. Recibe io.Reader e io.Writer para poder probarlo.
package consola

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"asistente/internal/asistente"
)

// TiempoMaximo por pregunta, contando búsqueda, modelo y reintentos.
const TiempoMaximo = 3 * time.Minute

const ayuda = `Escribe una pregunta y pulsa Intro.
Comandos: /nueva (olvida la conversación), /ayuda, /salir.`

// Ejecutar es el bucle interactivo. Termina con /salir, con el fin de la
// entrada (Ctrl+D, o Ctrl+Z e Intro en Windows) o cuando se cancela ctx (Ctrl+C).
func Ejecutar(ctx context.Context, entrada io.Reader, salida io.Writer, conv *asistente.Conversacion) error {
	lector := bufio.NewScanner(entrada)
	lector.Buffer(make([]byte, 64*1024), 64*1024)
	fmt.Fprintln(salida, ayuda)
	for {
		fmt.Fprint(salida, "\n> ")
		if !lector.Scan() {
			fmt.Fprintln(salida)
			return lector.Err()
		}
		linea := strings.TrimSpace(lector.Text())
		switch linea {
		case "":
			continue
		case "/salir", "/exit":
			return nil
		case "/ayuda":
			fmt.Fprintln(salida, ayuda)
			continue
		case "/nueva":
			conv.Olvidar()
			fmt.Fprintln(salida, "Conversación nueva.")
			continue
		}
		if err := Preguntar(ctx, salida, conv, linea); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			fmt.Fprintln(salida, "Error:", err)
		}
	}
}

// Preguntar hace una sola pregunta y escribe la respuesta con sus fuentes.
func Preguntar(ctx context.Context, salida io.Writer, conv *asistente.Conversacion, pregunta string) error {
	ctx, cancelar := context.WithTimeout(ctx, TiempoMaximo)
	defer cancelar()

	fmt.Fprintln(salida, "Buscando y redactando...")
	r, err := conv.Responder(ctx, pregunta)
	if err != nil {
		return err
	}
	if r.Aviso != "" {
		fmt.Fprintln(salida, "Aviso:", r.Aviso)
	}
	fmt.Fprintln(salida)
	fmt.Fprintln(salida, r.Texto)
	if len(r.Fuentes) > 0 {
		fmt.Fprintln(salida, "\nFuentes:")
		for i, f := range r.Fuentes {
			fmt.Fprintf(salida, "  [%d] %s\n      %s\n", i+1, f.Titulo, f.URL)
		}
	}
	return nil
}
