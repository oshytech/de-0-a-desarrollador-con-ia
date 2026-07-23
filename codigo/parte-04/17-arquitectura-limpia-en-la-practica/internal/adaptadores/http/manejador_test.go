package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"reservas/internal/adaptadores/memoria"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
)

func nuevoServidorDePrueba() http.Handler {
	hoy := time.Date(2026, 12, 1, 9, 0, 0, 0, time.UTC)
	socios, reservas := memoria.DatosDeEjemplo(hoy)
	listar := aplicacion.NuevoListarReservasDeSocio(socios, reservas, reloj.Fijo{Momento: hoy})
	return NuevoManejador(CasosDeUso{Listar: listar}, time.UTC).Rutas()
}

func TestListarDevuelveJSON(t *testing.T) {
	respuesta := httptest.NewRecorder()
	nuevoServidorDePrueba().ServeHTTP(respuesta, httptest.NewRequest("GET", "/socios/s-1/reservas", nil))

	if respuesta.Code != http.StatusOK {
		t.Fatalf("esperaba 200, salió %d", respuesta.Code)
	}
	var cuerpo []reservaJSON
	if err := json.NewDecoder(respuesta.Body).Decode(&cuerpo); err != nil {
		t.Fatal(err)
	}
	if len(cuerpo) != 2 || cuerpo[0].Inicio != "2026-12-02T10:00:00Z" {
		t.Fatalf("respuesta inesperada: %+v", cuerpo)
	}
}

func TestSocioQueNoExisteDa404(t *testing.T) {
	respuesta := httptest.NewRecorder()
	nuevoServidorDePrueba().ServeHTTP(respuesta, httptest.NewRequest("GET", "/socios/s-99/reservas", nil))
	if respuesta.Code != http.StatusNotFound {
		t.Fatalf("esperaba 404, salió %d", respuesta.Code)
	}
}
