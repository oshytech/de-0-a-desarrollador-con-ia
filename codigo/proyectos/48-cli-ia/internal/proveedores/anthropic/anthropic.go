// Package anthropic adapta la API de mensajes de Anthropic al puerto Modelo.
// Usa net/http directamente para que se vea qué viaja por la red.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"asistente/internal/asistente"
	"asistente/internal/red"
)

const (
	URLPorDefecto    = "https://api.anthropic.com"
	ModeloPorDefecto = "claude-opus-5-5"
	versionAPI       = "2026-10-21"
)

// Modelo habla con la API de Anthropic.
type Modelo struct {
	clave   string
	modelo  string
	url     string
	cliente *http.Client
}

func Nuevo(clave, modelo, url string) *Modelo {
	if modelo == "" {
		modelo = ModeloPorDefecto
	}
	if url == "" {
		url = URLPorDefecto
	}
	return &Modelo{clave: clave, modelo: modelo, url: url, cliente: &http.Client{Timeout: 2 * time.Minute}}
}

type mensaje struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type peticion struct {
	Model        string            `json:"model"`
	MaxTokens    int               `json:"max_tokens"`
	System       string            `json:"system"`
	Messages     []mensaje         `json:"messages"`
	OutputConfig map[string]string `json:"output_config"`
	Fallbacks    string            `json:"fallbacks"`
}

type respuesta struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

var ErrRechazo = errors.New("el modelo ha rechazado la petición")

func (m *Modelo) Completar(ctx context.Context, sistema string, mensajes []asistente.Mensaje) (string, error) {
	cuerpo := peticion{
		Model:        m.modelo,
		MaxTokens:    16000,
		System:       sistema,
		OutputConfig: map[string]string{"effort": "low"}, // preguntas y respuestas: poco razonamiento basta
		Fallbacks:    "default",                          // si el modelo rechaza, la API prueba con otro
	}
	for _, msj := range mensajes {
		cuerpo.Messages = append(cuerpo.Messages, mensaje{Role: string(msj.Rol), Content: msj.Texto})
	}
	datos, err := json.Marshal(cuerpo)
	if err != nil {
		return "", err
	}

	crudo, err := red.Enviar(ctx, m.cliente, red.PoliticaInteractiva, func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodPost, m.url+"/v1/messages", bytes.NewReader(datos))
		if err != nil {
			return nil, err
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("x-api-key", m.clave)
		req.Header.Set("anthropic-version", versionAPI)
		req.Header.Set("anthropic-beta", "server-side-fallback-2026-07-01")
		return req, nil
	})
	if err != nil {
		return "", err
	}

	var r respuesta
	if err := json.Unmarshal(crudo, &r); err != nil {
		return "", fmt.Errorf("respuesta de Anthropic ilegible: %w", err)
	}
	if r.StopReason == "refusal" {
		return "", ErrRechazo
	}
	var texto bytes.Buffer
	for _, bloque := range r.Content {
		if bloque.Type == "text" { // los bloques de razonamiento se ignoran
			texto.WriteString(bloque.Text)
		}
	}
	if r.StopReason == "max_tokens" {
		texto.WriteString("\n[respuesta cortada por longitud]")
	}
	return texto.String(), nil
}
