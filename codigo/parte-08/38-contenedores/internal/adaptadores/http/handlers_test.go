package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	apihttp "reservas/internal/adaptadores/http"
	"reservas/internal/dominio"
)

type reservadorFalso struct{ err error }

func (f reservadorFalso) Reservar(context.Context, int64, int64, time.Time) (int64, error) {
	return 7, f.err
}

type bdFalsa struct{ err error }

func (b bdFalsa) Ping(context.Context) error { return b.err }

func TestCodigosHTTP(t *testing.T) {
	cuerpo := `{"socio_id":1,"pista_id":1,"inicio":"2026-12-01T10:00:00+01:00"}`
	casos := []struct {
		nombre string
		err    error
		quiero int
	}{
		{"confirmada", nil, http.StatusCreated},
		{"pista ocupada", dominio.ErrPistaOcupada, http.StatusConflict},
		{"límite", dominio.ErrLimiteReservas, http.StatusUnprocessableEntity},
		{"error inesperado", errors.New("se cayó la red"), http.StatusInternalServerError},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			mux := apihttp.NuevoMux(reservadorFalso{c.err}, bdFalsa{})
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest("POST", "/reservas", strings.NewReader(cuerpo)))
			if rec.Code != c.quiero {
				t.Fatalf("código %d, quería %d", rec.Code, c.quiero)
			}
		})
	}
}

func TestVivoNoDependeDeLaBaseDeDatos(t *testing.T) {
	mux := apihttp.NuevoMux(reservadorFalso{}, bdFalsa{errors.New("caída")})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/vivo", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/vivo debe responder 200 aunque la base de datos falle, dio %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/listo", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("/listo debe responder 503 si la base de datos falla, dio %d", rec.Code)
	}
}
