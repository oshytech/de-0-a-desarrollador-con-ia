package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"reservas/internal/adaptadores/eventos"
	"reservas/internal/adaptadores/memoria"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
)

// Tests de aceptación: atraviesan el corte entero (HTTP, caso de uso,
// dominio y repositorios en memoria) con la hora fijada a las 9:30.
var ahora = time.Date(2026, 12, 1, 9, 30, 0, 0, time.UTC)

func nuevaAPI() http.Handler {
	socios, pistas, reservas := memoria.DatosDeEjemplo()
	relojFijo := reloj.Fijo{Momento: ahora}
	casos := CasosDeUso{
		Reservar: aplicacion.NuevoReservarPista(reservas, socios, pistas, relojFijo, eventos.NuevoBus(1, 0)),
		Cancelar: aplicacion.NuevoCancelarReserva(reservas, relojFijo, eventos.NuevoBus(1, 0)),
		Listar:   aplicacion.NuevoListarReservasDeSocio(socios, reservas, relojFijo),
	}
	return NuevoManejador(casos, time.UTC).Rutas()
}

func pedir(t *testing.T, api http.Handler, metodo, ruta, cuerpo string) (int, map[string]any) {
	t.Helper()
	respuesta := httptest.NewRecorder()
	api.ServeHTTP(respuesta, httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo)))
	var datos map[string]any
	json.Unmarshal(respuesta.Body.Bytes(), &datos)
	return respuesta.Code, datos
}

func TestReservarUnaPistaLibre(t *testing.T) {
	estado, datos := pedir(t, nuevaAPI(), "POST", "/reservas",
		`{"socio_id":"s-1","pista_id":"p-1","inicio":"2026-12-02T10:00"}`)
	if estado != http.StatusCreated || datos["estado"] != "confirmada" || datos["fin"] != "2026-12-02T11:00:00Z" {
		t.Fatalf("esperaba 201 y una reserva confirmada, salió %d %v", estado, datos)
	}
}

func TestLaMismaHoraNoSeReservaDosVeces(t *testing.T) {
	api := nuevaAPI()
	pedir(t, api, "POST", "/reservas", `{"socio_id":"s-1","pista_id":"p-1","inicio":"2026-12-02T10:00"}`)
	estado, datos := pedir(t, api, "POST", "/reservas", `{"socio_id":"s-3","pista_id":"p-1","inicio":"2026-12-02T10:00"}`)
	if estado != http.StatusConflict || datos["error"] != "pista_ocupada" {
		t.Fatalf("esperaba 409 pista_ocupada, salió %d %v", estado, datos)
	}
}

func TestRespuestasDeError(t *testing.T) {
	casos := []struct {
		nombre, cuerpo string
		estado         int
		codigo         string
	}{
		{"JSON roto", `{"socio_id":`, 400, "peticion_no_valida"},
		{"fecha mal escrita", `{"socio_id":"s-1","pista_id":"p-1","inicio":"mañana"}`, 400, "peticion_no_valida"},
		{"no empieza en punto", `{"socio_id":"s-1","pista_id":"p-1","inicio":"2026-12-02T10:30"}`, 422, "fuera_de_horario"},
		{"socio de baja", `{"socio_id":"s-2","pista_id":"p-1","inicio":"2026-12-02T10:00"}`, 422, "socio_inactivo"},
		{"pista que no existe", `{"socio_id":"s-1","pista_id":"p-9","inicio":"2026-12-02T10:00"}`, 404, "no_encontrado"},
		{"dentro de un mes", `{"socio_id":"s-1","pista_id":"p-1","inicio":"2027-01-02T10:00"}`, 422, "fuera_de_plazo"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			estado, datos := pedir(t, nuevaAPI(), "POST", "/reservas", c.cuerpo)
			if estado != c.estado || datos["error"] != c.codigo {
				t.Fatalf("esperaba %d %s, salió %d %v", c.estado, c.codigo, estado, datos)
			}
		})
	}
}

func TestCancelarYVolverAReservar(t *testing.T) {
	api := nuevaAPI()
	_, reserva := pedir(t, api, "POST", "/reservas", `{"socio_id":"s-1","pista_id":"p-1","inicio":"2026-12-02T10:00"}`)
	ruta := "/reservas/" + reserva["id"].(string) + "/cancelacion"

	if estado, datos := pedir(t, api, "POST", ruta, `{"socio_id":"s-3"}`); estado != http.StatusForbidden {
		t.Fatalf("otro socio no debería poder cancelar: %d %v", estado, datos)
	}
	if estado, datos := pedir(t, api, "POST", ruta, `{"socio_id":"s-1"}`); estado != http.StatusOK || datos["estado"] != "cancelada" {
		t.Fatalf("esperaba 200 cancelada, salió %d %v", estado, datos)
	}
	if estado, _ := pedir(t, api, "POST", ruta, `{"socio_id":"s-1"}`); estado != http.StatusConflict {
		t.Fatalf("cancelar dos veces debería dar 409, salió %d", estado)
	}
	if estado, _ := pedir(t, api, "POST", "/reservas", `{"socio_id":"s-3","pista_id":"p-1","inicio":"2026-12-02T10:00"}`); estado != http.StatusCreated {
		t.Fatalf("la hora cancelada debería quedar libre, salió %d", estado)
	}
}
