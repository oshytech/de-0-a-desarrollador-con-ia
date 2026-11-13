// Comando asistente: un asistente de terminal que responde con búsqueda web.
//
// Uso:
//
//	asistente                    modo interactivo
//	asistente -p "pregunta"      una sola pregunta y termina
//	asistente -simulado          sin red ni claves (respuestas de prueba)
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"asistente/internal/asistente"
	"asistente/internal/consola"
	"asistente/internal/proveedores/anthropic"
	"asistente/internal/proveedores/openai"
	"asistente/internal/proveedores/simulado"
	"asistente/internal/proveedores/tavily"
)

func main() {
	if err := ejecutar(os.Args[1:], os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func ejecutar(args []string, entorno func(string) string, entrada io.Reader, salida, errores io.Writer) error {
	banderas := flag.NewFlagSet("asistente", flag.ContinueOnError)
	banderas.SetOutput(errores)
	pregunta := banderas.String("p", "", "hace una sola pregunta y termina")
	forzarSimulado := banderas.Bool("simulado", false, "usa el modelo y el buscador simulados")
	if err := banderas.Parse(args); err != nil {
		return err
	}

	modelo, buscador, descripcion, err := configurar(entorno, *forzarSimulado)
	if err != nil {
		return err
	}
	fmt.Fprintln(errores, "Proveedor:", descripcion)

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer parar()

	conv := asistente.NuevaConversacion(modelo, buscador)
	if *pregunta != "" {
		return consola.Preguntar(ctx, salida, conv, *pregunta)
	}
	err = consola.Ejecutar(ctx, entrada, salida, conv)
	if errors.Is(err, context.Canceled) {
		return nil // Ctrl+C es una forma normal de salir
	}
	return err
}

// configurar elige los adaptadores según las variables de entorno.
// Es el único sitio del programa que conoce a los proveedores concretos.
func configurar(entorno func(string) string, forzarSimulado bool) (asistente.Modelo, asistente.Buscador, string, error) {
	proveedor := entorno("ASISTENTE_PROVEEDOR")
	if forzarSimulado {
		proveedor = "simulado"
	}
	if proveedor == "" {
		// Sin configuración explícita: Anthropic si hay clave, si no, el modo simulado.
		proveedor = "simulado"
		if entorno("ANTHROPIC_API_KEY") != "" {
			proveedor = "anthropic"
		}
	}
	if proveedor == "simulado" {
		return simulado.Modelo{}, simulado.Buscador{}, "simulado (sin red ni claves)", nil
	}

	claveBusqueda := entorno("TAVILY_API_KEY")
	if claveBusqueda == "" {
		return nil, nil, "", errors.New("falta TAVILY_API_KEY para la búsqueda web (o usa -simulado)")
	}
	buscador := tavily.Nuevo(claveBusqueda, "")
	modeloElegido := entorno("ASISTENTE_MODELO")

	switch proveedor {
	case "anthropic":
		clave := entorno("ANTHROPIC_API_KEY")
		if clave == "" {
			return nil, nil, "", errors.New("falta ANTHROPIC_API_KEY")
		}
		m := anthropic.Nuevo(clave, modeloElegido, "")
		return m, buscador, "anthropic", nil
	case "openai", "ollama":
		url := entorno("ASISTENTE_URL_BASE")
		if proveedor == "ollama" && url == "" {
			url = "http://localhost:11434/v1"
		}
		m, err := openai.Nuevo(entorno("OPENAI_API_KEY"), modeloElegido, url)
		if err != nil {
			return nil, nil, "", err
		}
		return m, buscador, proveedor + " (" + modeloElegido + ")", nil
	}
	return nil, nil, "", fmt.Errorf("proveedor desconocido %q: usa anthropic, openai, ollama o simulado", proveedor)
}
