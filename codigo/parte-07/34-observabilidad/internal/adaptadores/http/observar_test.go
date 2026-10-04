package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"reservas/internal/dominio"
	"reservas/internal/metricas"
)

func TestObservarPropagaElIdentificador(t *testing.T) {
	var salida bytes.Buffer
	m := metricas.NuevoRegistro()
	mux := http.NewServeMux()
	mux.Handle("POST /reservas", ManejadorReservar(func(context.Context, string, string, time.Time) error { return nil }))
	h := Observar(slog.New(slog.NewJSONHandler(&salida, nil)), m, mux)

	pet := httptest.NewRequest("POST", "/reservas", strings.NewReader(`{"socio":"s1","pista":"pista-1"}`))
	pet.Header.Set(CabeceraID, "abc123")
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, pet)

	if got := resp.Header().Get(CabeceraID); got != "abc123" {
		t.Fatalf("cabecera %s = %q, quería abc123", CabeceraID, got)
	}
	var linea map[string]any
	if err := json.Unmarshal(salida.Bytes(), &linea); err != nil {
		t.Fatalf("el log no es JSON: %v\n%s", err, salida.String())
	}
	if linea["request_id"] != "abc123" || linea["ruta"] != "POST /reservas" || linea["estado"] != float64(201) {
		t.Fatalf("línea de log inesperada: %v", linea)
	}
	if n := m.Contador("POST /reservas", 201); n != 1 {
		t.Fatalf("contador = %d, quería 1", n)
	}
}

func TestErroresDeNegocioYErroresDelSistema(t *testing.T) {
	casos := []struct {
		nombre     string
		err        error
		estado     int
		nivel      string
		mensajeLog string
	}{
		{"pista ocupada", dominio.ErrPistaOcupada, 409, "INFO", "reserva rechazada"},
		{"fuera de plazo", dominio.ErrFueraDePlazo, 422, "INFO", "reserva rechazada"},
		{"socio inactivo", dominio.ErrSocioInactivo, 403, "INFO", "reserva rechazada"},
		{"base de datos caída", fmt.Errorf("guardar reserva: %w", errors.New("conexión rechazada")), 500, "ERROR", "no se pudo reservar"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			var salida bytes.Buffer
			log := slog.New(slog.NewJSONHandler(&salida, nil))
			mux := http.NewServeMux()
			mux.Handle("POST /reservas", ManejadorReservar(func(context.Context, string, string, time.Time) error { return c.err }))
			h := Observar(log, metricas.NuevoRegistro(), mux)

			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, httptest.NewRequest("POST", "/reservas", strings.NewReader(`{"socio":"s1","pista":"pista-1"}`)))

			if resp.Code != c.estado {
				t.Fatalf("estado = %d, quería %d", resp.Code, c.estado)
			}
			primera := strings.SplitN(salida.String(), "\n", 2)[0]
			var linea map[string]any
			_ = json.Unmarshal([]byte(primera), &linea)
			if linea["level"] != c.nivel || linea["msg"] != c.mensajeLog {
				t.Fatalf("primera línea de log = %s", primera)
			}
		})
	}
}

func TestSalud(t *testing.T) {
	casos := []struct {
		nombre string
		bd     error
		estado int
	}{
		{"todo bien", nil, 200},
		{"base de datos caída", errors.New("timeout"), 503},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			h := ManejadorSalud(map[string]Comprobacion{
				"base_de_datos": func(context.Context) error { return c.bd },
			})
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, httptest.NewRequest("GET", "/health", nil))
			if resp.Code != c.estado {
				t.Fatalf("estado = %d, quería %d (%s)", resp.Code, c.estado, resp.Body.String())
			}
		})
	}
}
