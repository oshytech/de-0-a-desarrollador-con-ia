package http

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	_ "time/tzdata"
)

func nuevoManejador(t *testing.T) ManejadorReservar {
	t.Helper()
	zona, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		t.Fatal(err)
	}
	// "Ahora" fijo: lunes 30 de noviembre de 2026 a las 10:15 en Madrid.
	ahora := time.Date(2026, 11, 30, 10, 15, 0, 0, zona)
	return ManejadorReservar{
		Ahora:    func() time.Time { return ahora },
		ZonaClub: zona,
		Log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestReservar(t *testing.T) {
	casos := []struct {
		nombre string
		cuerpo string
		tipo   string
		estado int
	}{
		{"válida", `{"pista_id":"padel-2","inicio":"2026-12-01T18:00:00+01:00"}`, "application/json", http.StatusAccepted},
		{"sin JSON", `pista=padel-2`, "application/x-www-form-urlencoded", http.StatusBadRequest},
		{"campo extra que intenta elegir socio", `{"pista_id":"padel-2","inicio":"2026-12-01T18:00:00+01:00","socio_id":7}`, "application/json", http.StatusBadRequest},
		{"dos objetos seguidos", `{"pista_id":"padel-2","inicio":"2026-12-01T18:00:00+01:00"}{}`, "application/json", http.StatusBadRequest},
		{"pista con comilla", `{"pista_id":"padel-2' OR '1'='1","inicio":"2026-12-01T18:00:00+01:00"}`, "application/json", http.StatusBadRequest},
		{"pista inexistente", `{"pista_id":"golf-1","inicio":"2026-12-01T18:00:00+01:00"}`, "application/json", http.StatusBadRequest},
		{"fecha mal escrita", `{"pista_id":"tenis-1","inicio":"mañana a las seis"}`, "application/json", http.StatusBadRequest},
		{"no en punto", `{"pista_id":"tenis-1","inicio":"2026-12-01T18:30:00+01:00"}`, "application/json", http.StatusUnprocessableEntity},
		{"de madrugada", `{"pista_id":"tenis-1","inicio":"2026-12-01T03:00:00+01:00"}`, "application/json", http.StatusUnprocessableEntity},
		{"en el pasado", `{"pista_id":"tenis-1","inicio":"2026-11-29T18:00:00+01:00"}`, "application/json", http.StatusUnprocessableEntity},
		{"dentro de un mes", `{"pista_id":"tenis-1","inicio":"2026-12-30T18:00:00+01:00"}`, "application/json", http.StatusUnprocessableEntity},
		{"cuerpo enorme", `{"pista_id":"tenis-1","inicio":"` + strings.Repeat("9", 5000) + `"}`, "application/json", http.StatusRequestEntityTooLarge},
	}

	m := nuevoManejador(t)
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/reservas", strings.NewReader(c.cuerpo))
			req.Header.Set("Content-Type", c.tipo)
			rec := httptest.NewRecorder()
			m.ServeHTTP(rec, req)
			if rec.Code != c.estado {
				t.Fatalf("estado %d, esperaba %d (respuesta: %s)", rec.Code, c.estado, rec.Body.String())
			}
		})
	}
}

func TestLaHoraSeMideEnLaZonaDelClub(t *testing.T) {
	// Las 20:00 en UTC son las 21:00 en Madrid en invierno: es la última
	// franja válida. Las 21:00 UTC son las 22:00 en Madrid: fuera de horario.
	m := nuevoManejador(t)
	for inicio, estado := range map[string]int{
		"2026-12-01T20:00:00Z": http.StatusAccepted,
		"2026-12-01T21:00:00Z": http.StatusUnprocessableEntity,
	} {
		cuerpo := `{"pista_id":"tenis-4","inicio":"` + inicio + `"}`
		req := httptest.NewRequest(http.MethodPost, "/reservas", strings.NewReader(cuerpo))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		m.ServeHTTP(rec, req)
		if rec.Code != estado {
			t.Errorf("%s: estado %d, esperaba %d", inicio, rec.Code, estado)
		}
	}
}
