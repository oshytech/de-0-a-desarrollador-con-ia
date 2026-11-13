// Tests de los adaptadores contra servidores falsos: comprueban que se envía
// lo que espera cada API y que se interpreta bien lo que devuelve, sin red ni claves.
package proveedores_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"asistente/internal/asistente"
	"asistente/internal/proveedores/anthropic"
	"asistente/internal/proveedores/openai"
	"asistente/internal/proveedores/tavily"
)

var conversacion = []asistente.Mensaje{
	{Rol: asistente.Usuario, Texto: "hola"},
	{Rol: asistente.Asistente, Texto: "hola, ¿qué necesitas?"},
	{Rol: asistente.Usuario, Texto: "¿cuánto mide una pista de pádel?"},
}

// falso levanta un servidor que comprueba la petición y responde con cuerpo.
func falso(t *testing.T, ruta string, comprobar func(*http.Request, map[string]any), cuerpo string) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != ruta || r.Method != http.MethodPost {
			t.Errorf("petición a %s %s, quiero POST %s", r.Method, r.URL.Path, ruta)
		}
		datos, _ := io.ReadAll(r.Body)
		var cuerpoJSON map[string]any
		if err := json.Unmarshal(datos, &cuerpoJSON); err != nil {
			t.Errorf("cuerpo no es JSON: %s", datos)
		}
		comprobar(r, cuerpoJSON)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, cuerpo)
	}))
	t.Cleanup(s.Close)
	return s
}

func TestAnthropic(t *testing.T) {
	s := falso(t, "/v1/messages", func(r *http.Request, c map[string]any) {
		if r.Header.Get("x-api-key") != "clave-de-mentira" || r.Header.Get("anthropic-version") == "" {
			t.Errorf("cabeceras incorrectas: %v", r.Header)
		}
		if c["model"] != "claude-opus-5-5" || c["system"] != "sé breve" || len(c["messages"].([]any)) != 3 {
			t.Errorf("cuerpo incorrecto: %v", c)
		}
	}, `{"content":[{"type":"thinking","thinking":""},{"type":"text","text":"Mide 20 por 10 metros [1]."}],"stop_reason":"end_turn"}`)

	texto, err := anthropic.Nuevo("clave-de-mentira", "", s.URL).Completar(context.Background(), "sé breve", conversacion)
	if err != nil || texto != "Mide 20 por 10 metros [1]." {
		t.Fatalf("texto %q, error %v", texto, err)
	}
}

func TestAnthropicRechazo(t *testing.T) {
	s := falso(t, "/v1/messages", func(*http.Request, map[string]any) {},
		`{"content":[],"stop_reason":"refusal"}`)
	_, err := anthropic.Nuevo("k", "", s.URL).Completar(context.Background(), "", conversacion)
	if !errors.Is(err, anthropic.ErrRechazo) {
		t.Fatalf("error = %v, quiero ErrRechazo", err)
	}
}

func TestOpenAICompatible(t *testing.T) {
	s := falso(t, "/v1/chat/completions", func(r *http.Request, c map[string]any) {
		if r.Header.Get("Authorization") != "" {
			t.Error("sin clave (Ollama) no debe enviarse Authorization")
		}
		mensajes := c["messages"].([]any)
		primero := mensajes[0].(map[string]any)
		if c["model"] != "llama3.1" || len(mensajes) != 4 || primero["role"] != "system" {
			t.Errorf("cuerpo incorrecto: %v", c)
		}
	}, `{"choices":[{"message":{"role":"assistant","content":"20 x 10 m"}}]}`)

	m, err := openai.Nuevo("", "llama3.1", s.URL+"/v1")
	if err != nil {
		t.Fatal(err)
	}
	texto, err := m.Completar(context.Background(), "sé breve", conversacion)
	if err != nil || texto != "20 x 10 m" {
		t.Fatalf("texto %q, error %v", texto, err)
	}
	if _, err := openai.Nuevo("", "", ""); err == nil {
		t.Fatal("sin modelo debería fallar")
	}
}

func TestTavily(t *testing.T) {
	s := falso(t, "/search", func(r *http.Request, c map[string]any) {
		if r.Header.Get("Authorization") != "Bearer tvly-de-mentira" || c["query"] != "pádel" || c["max_results"] != 5.0 {
			t.Errorf("petición incorrecta: %v %v", r.Header, c)
		}
	}, `{"results":[{"title":"Pádel","url":"https://ejemplo.test","content":"20 por 10","score":0.9}]}`)

	fuentes, err := tavily.Nuevo("tvly-de-mentira", s.URL).Buscar(context.Background(), "pádel", 5)
	if err != nil || len(fuentes) != 1 || fuentes[0].URL != "https://ejemplo.test" {
		t.Fatalf("fuentes %+v, error %v", fuentes, err)
	}
}
