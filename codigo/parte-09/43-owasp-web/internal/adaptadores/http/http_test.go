package http

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"owaspweb/internal/adaptadores/memoria"
	"owaspweb/internal/aplicacion"
	"owaspweb/internal/dominio"
)

type relojFijo time.Time

func (r relojFijo) Ahora() time.Time { return time.Time(r) }

var ahora = time.Date(2026, 11, 30, 10, 0, 0, 0, time.UTC)

const (
	ana   int64 = 1
	bruno int64 = 2
)

// Ana tiene la reserva 1043 mañana. Bruno es otro socio con sesión válida.
func escenario() (*memoria.Reservas, Sesiones) {
	repo := memoria.NuevasReservas(dominio.Reserva{
		ID: 1043, SocioID: ana, PistaID: "padel-2",
		Inicio: ahora.Add(24 * time.Hour), Estado: dominio.Confirmada,
	})
	return repo, Sesiones{"token-ana": ana, "token-bruno": bruno}
}

func servidor(c Canceladora, s Sesiones, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("DELETE /reservas/{id}", s.Identificar(CancelarReserva(c, log)))
	return mux
}

func cancelar(h http.Handler, token, id string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/reservas/"+id, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func estadoDe(t *testing.T, repo *memoria.Reservas, id int64) dominio.Estado {
	t.Helper()
	r, err := repo.Buscar(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return r.Estado
}

var silencio = slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

// Este test documenta el fallo: con la versión vulnerable, Bruno cancela la
// reserva de Ana y el servidor responde 204 como si todo fuera bien.
func TestVersionVulnerable_BrunoCancelaLaReservaDeAna(t *testing.T) {
	repo, sesiones := escenario()
	h := servidor(aplicacion.CancelarReservaVulnerable{Reservas: repo, Reloj: relojFijo(ahora)}, sesiones, silencio)

	rec := cancelar(h, "token-bruno", "1043")

	if rec.Code != http.StatusNoContent || estadoDe(t, repo, 1043) != dominio.Cancelada {
		t.Fatalf("esperaba demostrar el fallo, pero obtuve %d", rec.Code)
	}
}

func TestCancelarReserva_ControlDeAcceso(t *testing.T) {
	casos := []struct {
		nombre string
		token  string
		id     string
		estado int
	}{
		{"sin sesión", "", "1043", http.StatusUnauthorized},
		{"token inventado", "token-falso", "1043", http.StatusUnauthorized},
		{"Bruno intenta cancelar la de Ana", "token-bruno", "1043", http.StatusNotFound},
		{"Bruno prueba un id que no existe", "token-bruno", "9999", http.StatusNotFound},
		{"id que no es un número", "token-ana", "1043%20OR%201=1", http.StatusBadRequest},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo, sesiones := escenario()
			h := servidor(aplicacion.CancelarReserva{Reservas: repo, Reloj: relojFijo(ahora)}, sesiones, silencio)
			rec := cancelar(h, c.token, c.id)
			if rec.Code != c.estado {
				t.Fatalf("estado %d, esperaba %d", rec.Code, c.estado)
			}
			if estadoDe(t, repo, 1043) != dominio.Confirmada {
				t.Fatal("la reserva de Ana no debería haber cambiado")
			}
		})
	}
}

func TestCancelarReserva_LaDueñaSíPuede(t *testing.T) {
	repo, sesiones := escenario()
	h := servidor(aplicacion.CancelarReserva{Reservas: repo, Reloj: relojFijo(ahora)}, sesiones, silencio)

	if rec := cancelar(h, "token-ana", "1043"); rec.Code != http.StatusNoContent {
		t.Fatalf("estado %d, esperaba 204", rec.Code)
	}
	if estadoDe(t, repo, 1043) != dominio.Cancelada {
		t.Fatal("la reserva debería estar cancelada")
	}
}

// Un fallo interno no debe enseñar detalles al cliente (A10).
type repoRoto struct{}

func (repoRoto) Buscar(context.Context, int64) (dominio.Reserva, error) {
	return dominio.Reserva{}, errors.New(`pq: relation "reservas" does not exist at /srv/club/repo.go:42`)
}
func (repoRoto) Guardar(context.Context, dominio.Reserva) error { return nil }

func TestErrorInterno_RespuestaGenericaYDetalleEnElLog(t *testing.T) {
	var log bytes.Buffer
	_, sesiones := escenario()
	h := servidor(aplicacion.CancelarReserva{Reservas: repoRoto{}, Reloj: relojFijo(ahora)}, sesiones, slog.New(slog.NewJSONHandler(&log, nil)))

	rec := cancelar(h, "token-ana", "1043")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("estado %d, esperaba 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "relation") || strings.Contains(rec.Body.String(), "/srv/") {
		t.Fatalf("la respuesta filtra detalles internos: %s", rec.Body.String())
	}
	if !strings.Contains(log.String(), "relation") {
		t.Fatal("el detalle tendría que estar en el log del servidor")
	}
}
