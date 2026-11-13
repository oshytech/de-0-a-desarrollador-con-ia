package consola_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"asistente/internal/asistente"
	"asistente/internal/consola"
	"asistente/internal/proveedores/simulado"
)

// TestSesionSimulada recorre una sesión completa en modo simulado: sin red ni claves.
func TestSesionSimulada(t *testing.T) {
	entrada := strings.NewReader("¿Cuánto mide una pista de pádel?\n¿Y en tenis?\n/nueva\nOtra pregunta sobre pádel\n/salir\nesto ya no se lee\n")
	var salida bytes.Buffer
	conv := asistente.NuevaConversacion(simulado.Modelo{}, simulado.Buscador{})

	if err := consola.Ejecutar(context.Background(), entrada, &salida, conv); err != nil {
		t.Fatal(err)
	}
	texto := salida.String()
	for _, trozo := range []string{
		"[1] Reglamento de pádel: dimensiones de la pista",
		"https://ejemplo.test/padel/pista",
		"Turnos anteriores que recuerdo: 1.",
		"Conversación nueva.",
	} {
		if !strings.Contains(texto, trozo) {
			t.Errorf("la salida no contiene %q", trozo)
		}
	}
	if strings.Count(texto, "Turnos anteriores que recuerdo: 0.") != 2 {
		t.Errorf("tras /nueva la memoria debería empezar de cero:\n%s", texto)
	}
	if strings.Contains(texto, "esto ya no se lee") {
		t.Error("después de /salir no se procesa nada")
	}
}

func TestFinDeEntradaTerminaSinError(t *testing.T) {
	var salida bytes.Buffer
	conv := asistente.NuevaConversacion(simulado.Modelo{}, simulado.Buscador{})
	if err := consola.Ejecutar(context.Background(), strings.NewReader("hola\n"), &salida, conv); err != nil {
		t.Fatal(err)
	}
}

func TestErrorNoCortaLaSesion(t *testing.T) {
	larga := strings.Repeat("a", asistente.MaxCaracteresPreg+1)
	var salida bytes.Buffer
	conv := asistente.NuevaConversacion(simulado.Modelo{}, simulado.Buscador{})
	err := consola.Ejecutar(context.Background(), strings.NewReader(larga+"\ntenis\n/salir\n"), &salida, conv)
	if err != nil || !strings.Contains(salida.String(), "Error: la pregunta supera") ||
		!strings.Contains(salida.String(), "Tenis: cómo se cuentan los puntos") {
		t.Fatalf("error %v, salida:\n%s", err, salida.String())
	}
}
