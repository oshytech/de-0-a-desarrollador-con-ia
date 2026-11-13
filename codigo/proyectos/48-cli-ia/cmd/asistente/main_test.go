package main

import (
	"bytes"
	"strings"
	"testing"
)

// entornoDe simula variables de entorno sin tocar las reales.
func entornoDe(vars map[string]string) func(string) string {
	return func(nombre string) string { return vars[nombre] }
}

func TestConfiguracion(t *testing.T) {
	casos := []struct {
		nombre     string
		vars       map[string]string
		simulado   bool
		quiero     string // descripción esperada, o texto del error
		debeFallar bool   // true si debe fallar
	}{
		{"sin nada, modo simulado", nil, false, "simulado", false},
		{"la bandera manda", map[string]string{"ANTHROPIC_API_KEY": "k", "TAVILY_API_KEY": "t"}, true, "simulado", false},
		{"Anthropic por defecto si hay clave", map[string]string{"ANTHROPIC_API_KEY": "k", "TAVILY_API_KEY": "t"}, false, "anthropic", false},
		{"falta la clave de búsqueda", map[string]string{"ANTHROPIC_API_KEY": "k"}, false, "TAVILY_API_KEY", true},
		{"Ollama sin clave", map[string]string{"ASISTENTE_PROVEEDOR": "ollama", "ASISTENTE_MODELO": "llama3.1", "TAVILY_API_KEY": "t"}, false, "ollama (llama3.1)", false},
		{"OpenAI sin modelo", map[string]string{"ASISTENTE_PROVEEDOR": "openai", "TAVILY_API_KEY": "t"}, false, "ASISTENTE_MODELO", true},
		{"proveedor desconocido", map[string]string{"ASISTENTE_PROVEEDOR": "otro", "TAVILY_API_KEY": "t"}, false, "desconocido", true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, _, descripcion, err := configurar(entornoDe(c.vars), c.simulado)
			if c.debeFallar {
				if err == nil || !strings.Contains(err.Error(), c.quiero) {
					t.Fatalf("error = %v, quiero uno que mencione %q", err, c.quiero)
				}
				return
			}
			if err != nil || !strings.HasPrefix(descripcion, c.quiero) {
				t.Fatalf("descripción %q, error %v; quiero %q", descripcion, err, c.quiero)
			}
		})
	}
}

func TestUnaSolaPreguntaEnModoSimulado(t *testing.T) {
	var salida, errores bytes.Buffer
	err := ejecutar([]string{"-simulado", "-p", "¿cuántos puntos tiene un juego de tenis?"},
		entornoDe(nil), strings.NewReader(""), &salida, &errores)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(salida.String(), "Tenis: cómo se cuentan los puntos") {
		t.Fatalf("salida inesperada:\n%s", salida.String())
	}
}
