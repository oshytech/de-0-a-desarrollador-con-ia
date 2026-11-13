// Package openai adapta cualquier API compatible con el formato de chat de
// OpenAI al puerto Modelo. Sirve para OpenAI y para Ollama en local
// (http://localhost:11434/v1), que habla el mismo formato.
package openai

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

const URLPorDefecto = "https://api.openai.com/v1"

// Modelo habla con un endpoint /chat/completions.
type Modelo struct {
	clave   string // vacía en Ollama
	modelo  string
	url     string
	cliente *http.Client
}

func Nuevo(clave, modelo, url string) (*Modelo, error) {
	if modelo == "" {
		return nil, errors.New("falta el nombre del modelo (ASISTENTE_MODELO)")
	}
	if url == "" {
		url = URLPorDefecto
	}
	// Un modelo local puede tardar bastante más que uno en la nube.
	return &Modelo{clave: clave, modelo: modelo, url: url, cliente: &http.Client{Timeout: 5 * time.Minute}}, nil
}

type mensaje struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type peticion struct {
	Model    string    `json:"model"`
	Messages []mensaje `json:"messages"`
}

type respuesta struct {
	Choices []struct {
		Message mensaje `json:"message"`
	} `json:"choices"`
}

func (m *Modelo) Completar(ctx context.Context, sistema string, mensajes []asistente.Mensaje) (string, error) {
	cuerpo := peticion{Model: m.modelo, Messages: []mensaje{{Role: "system", Content: sistema}}}
	for _, msj := range mensajes {
		cuerpo.Messages = append(cuerpo.Messages, mensaje{Role: string(msj.Rol), Content: msj.Texto})
	}
	datos, err := json.Marshal(cuerpo)
	if err != nil {
		return "", err
	}
	crudo, err := red.Enviar(ctx, m.cliente, red.PoliticaInteractiva, func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodPost, m.url+"/chat/completions", bytes.NewReader(datos))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		if m.clave != "" {
			req.Header.Set("Authorization", "Bearer "+m.clave)
		}
		return req, nil
	})
	if err != nil {
		return "", err
	}
	var r respuesta
	if err := json.Unmarshal(crudo, &r); err != nil {
		return "", fmt.Errorf("respuesta ilegible: %w", err)
	}
	if len(r.Choices) == 0 {
		return "", errors.New("la respuesta no trae ninguna opción")
	}
	return r.Choices[0].Message.Content, nil
}
