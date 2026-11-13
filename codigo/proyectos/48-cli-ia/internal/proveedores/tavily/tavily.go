// Package tavily adapta la API de búsqueda de Tavily al puerto Buscador.
// La API es tan sencilla que no hace falta ninguna librería: una petición POST.
package tavily

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"asistente/internal/asistente"
	"asistente/internal/red"
)

const URLPorDefecto = "https://api.tavily.com"

type Buscador struct {
	clave   string
	url     string
	cliente *http.Client
}

func Nuevo(clave, url string) *Buscador {
	if url == "" {
		url = URLPorDefecto
	}
	return &Buscador{clave: clave, url: url, cliente: &http.Client{Timeout: 30 * time.Second}}
}

type peticion struct {
	Query       string `json:"query"`
	MaxResults  int    `json:"max_results"`
	SearchDepth string `json:"search_depth"`
}

type respuesta struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func (b *Buscador) Buscar(ctx context.Context, consulta string, maximo int) ([]asistente.Fuente, error) {
	datos, err := json.Marshal(peticion{Query: consulta, MaxResults: maximo, SearchDepth: "basic"})
	if err != nil {
		return nil, err
	}
	crudo, err := red.Enviar(ctx, b.cliente, red.PoliticaInteractiva, func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodPost, b.url+"/search", bytes.NewReader(datos))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+b.clave)
		return req, nil
	})
	if err != nil {
		return nil, err
	}
	var r respuesta
	if err := json.Unmarshal(crudo, &r); err != nil {
		return nil, fmt.Errorf("respuesta de búsqueda ilegible: %w", err)
	}
	fuentes := make([]asistente.Fuente, 0, len(r.Results))
	for _, res := range r.Results {
		fuentes = append(fuentes, asistente.Fuente{Titulo: res.Title, URL: res.URL, Contenido: res.Content})
	}
	return fuentes, nil
}
