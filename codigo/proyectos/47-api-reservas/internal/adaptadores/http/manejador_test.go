package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
	_ "time/tzdata"

	"reservas/internal/adaptadores/memoria"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
)

// Tests de aceptación: atraviesan el corte entero (HTTP, caso de uso,
// dominio y repositorios en memoria) con la hora fijada a las 9:30.
var ahora = time.Date(2026, 12, 1, 9, 30, 0, 0, time.UTC)

func nuevaAPI() http.Handler { return nuevaAPIEn(time.UTC) }

// nuevaAPIEn monta la API con la hora local del club en zona.
func nuevaAPIEn(zona *time.Location) http.Handler {
	socios, pistas, reservas := memoria.DatosDeEjemplo()
	relojFijo := reloj.Fijo{Momento: ahora}
	casos := CasosDeUso{
		Reservar: aplicacion.NuevoReservarPista(reservas, socios, pistas, relojFijo),
		Cancelar: aplicacion.NuevoCancelarReserva(reservas, relojFijo),
		Listar:   aplicacion.NuevoListarReservasDeSocio(socios, reservas, relojFijo),

		Pistas:         aplicacion.NuevoListarPistas(pistas),
		Disponibilidad: aplicacion.NuevoConsultarDisponibilidad(pistas, reservas, relojFijo),
	}
	return NuevoManejador(casos, zona).Rutas()
}

// pedir hace una petición en nombre de socio ("" para ir sin identidad).
func pedir(t *testing.T, api http.Handler, metodo, ruta, socio, cuerpo string) (int, map[string]any) {
	t.Helper()
	respuesta := httptest.NewRecorder()
	peticion := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if socio != "" {
		peticion.Header.Set(CabeceraSocio, socio)
	}
	api.ServeHTTP(respuesta, peticion)
	var datos map[string]any
	json.Unmarshal(respuesta.Body.Bytes(), &datos)
	return respuesta.Code, datos
}

func TestReservarUnaPistaLibre(t *testing.T) {
	estado, datos := pedir(t, nuevaAPI(), "POST", "/reservas",
		"s-1", `{"pista_id":"p-1","inicio":"2026-12-02T10:00"}`)
	if estado != http.StatusCreated || datos["estado"] != "confirmada" || datos["fin"] != "2026-12-02T11:00:00Z" {
		t.Fatalf("esperaba 201 y una reserva confirmada, salió %d %v", estado, datos)
	}
}

func TestLaMismaHoraNoSeReservaDosVeces(t *testing.T) {
	api := nuevaAPI()
	pedir(t, api, "POST", "/reservas", "s-1", `{"pista_id":"p-1","inicio":"2026-12-02T10:00"}`)
	estado, datos := pedir(t, api, "POST", "/reservas", "s-3", `{"pista_id":"p-1","inicio":"2026-12-02T10:00"}`)
	if estado != http.StatusConflict || datos["error"] != "pista_ocupada" {
		t.Fatalf("esperaba 409 pista_ocupada, salió %d %v", estado, datos)
	}
}

func TestRespuestasDeError(t *testing.T) {
	casos := []struct {
		nombre, socio, cuerpo string
		estado                int
		codigo                string
	}{
		{"sin identidad", "", `{"pista_id":"p-1","inicio":"2026-12-02T10:00"}`, 401, "no_autenticado"},
		{"JSON roto", "s-1", `{"pista_id":`, 400, "peticion_no_valida"},
		{"intenta reservar a nombre de otro", "s-1", `{"socio_id":"s-3","pista_id":"p-1","inicio":"2026-12-02T10:00"}`, 400, "peticion_no_valida"},
		{"fecha mal escrita", "s-1", `{"pista_id":"p-1","inicio":"mañana"}`, 400, "peticion_no_valida"},
		{"no empieza en punto", "s-1", `{"pista_id":"p-1","inicio":"2026-12-02T10:30"}`, 422, "fuera_de_horario"},
		{"socio de baja", "s-2", `{"pista_id":"p-1","inicio":"2026-12-02T10:00"}`, 422, "socio_inactivo"},
		{"pista que no existe", "s-1", `{"pista_id":"p-9","inicio":"2026-12-02T10:00"}`, 404, "no_encontrado"},
		{"dentro de un mes", "s-1", `{"pista_id":"p-1","inicio":"2027-01-02T10:00"}`, 422, "fuera_de_plazo"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			estado, datos := pedir(t, nuevaAPI(), "POST", "/reservas", c.socio, c.cuerpo)
			if estado != c.estado || datos["error"] != c.codigo {
				t.Fatalf("esperaba %d %s, salió %d %v", c.estado, c.codigo, estado, datos)
			}
		})
	}
}

func TestCancelarYVolverAReservar(t *testing.T) {
	api := nuevaAPI()
	_, reserva := pedir(t, api, "POST", "/reservas", "s-1", `{"pista_id":"p-1","inicio":"2026-12-02T10:00"}`)
	ruta := "/reservas/" + reserva["id"].(string) + "/cancelacion"

	// Para otro socio, la reserva sencillamente no existe.
	if estado, datos := pedir(t, api, "POST", ruta, "s-3", ""); estado != http.StatusNotFound {
		t.Fatalf("otro socio no debería poder cancelar ni saber que existe: %d %v", estado, datos)
	}
	if estado, datos := pedir(t, api, "POST", ruta, "s-1", ""); estado != http.StatusOK || datos["estado"] != "cancelada" {
		t.Fatalf("esperaba 200 cancelada, salió %d %v", estado, datos)
	}
	if estado, _ := pedir(t, api, "POST", ruta, "s-1", ""); estado != http.StatusConflict {
		t.Fatalf("cancelar dos veces debería dar 409, salió %d", estado)
	}
	if estado, _ := pedir(t, api, "POST", "/reservas", "s-3", `{"pista_id":"p-1","inicio":"2026-12-02T10:00"}`); estado != http.StatusCreated {
		t.Fatalf("la hora cancelada debería quedar libre, salió %d", estado)
	}
}

func TestCadaSocioSoloVeSusReservas(t *testing.T) {
	api := nuevaAPI()
	pedir(t, api, "POST", "/reservas", "s-1", `{"pista_id":"p-1","inicio":"2026-12-02T10:00"}`)

	estado, _ := pedir(t, api, "GET", "/socios/s-1/reservas", "s-1", "")
	if estado != http.StatusOK {
		t.Fatalf("el propio socio debería ver sus reservas, salió %d", estado)
	}
	estado, datos := pedir(t, api, "GET", "/socios/s-1/reservas", "s-3", "")
	if estado != http.StatusForbidden || datos["error"] != "prohibido" {
		t.Fatalf("otro socio no debería verlas, salió %d %v", estado, datos)
	}
}

func TestListarPistas(t *testing.T) {
	respuesta := httptest.NewRecorder()
	nuevaAPI().ServeHTTP(respuesta, httptest.NewRequest("GET", "/pistas", nil))
	var pistas []map[string]string
	json.Unmarshal(respuesta.Body.Bytes(), &pistas)
	if respuesta.Code != http.StatusOK || len(pistas) != 2 ||
		pistas[0]["id"] != "p-1" || pistas[0]["nombre"] != "Pista 1" || pistas[1]["deporte"] != "padel" {
		t.Fatalf("esperaba las dos pistas de ejemplo, salió %d %s", respuesta.Code, respuesta.Body)
	}
}

// horasLibres pide la disponibilidad sin identidad: es una consulta pública.
func horasLibres(t *testing.T, api http.Handler, ruta string) (int, string, []string) {
	t.Helper()
	respuesta := httptest.NewRecorder()
	api.ServeHTTP(respuesta, httptest.NewRequest("GET", ruta, nil))
	var datos struct {
		Fecha       string   `json:"fecha"`
		HorasLibres []string `json:"horas_libres"`
	}
	json.Unmarshal(respuesta.Body.Bytes(), &datos)
	return respuesta.Code, datos.Fecha, datos.HorasLibres
}

func TestLaHoraReservadaDejaDeEstarLibre(t *testing.T) {
	api := nuevaAPI()
	pedir(t, api, "POST", "/reservas", "s-1", `{"pista_id":"p-1","inicio":"2026-12-02T10:00"}`)

	estado, fecha, horas := horasLibres(t, api, "/pistas/p-1/disponibilidad?fecha=2026-12-02")
	if estado != http.StatusOK || fecha != "2026-12-02" || len(horas) != 13 || horas[0] != "08:00" || slices.Contains(horas, "10:00") {
		t.Fatalf("esperaba de 08:00 a 21:00 sin las 10:00, salió %d %s %v", estado, fecha, horas)
	}
	if _, _, horas := horasLibres(t, api, "/pistas/p-2/disponibilidad?fecha=2026-12-02"); len(horas) != 14 {
		t.Fatalf("la otra pista debería tener las 14 horas libres, salió %v", horas)
	}
}

// Con la hora del club en Madrid, las 9:30 UTC son las 10:30: la primera
// hora libre de hoy es la de las 11:00, en hora local.
func TestDisponibilidadEnHoraDelClub(t *testing.T) {
	madrid, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		t.Fatal(err)
	}
	_, _, horas := horasLibres(t, nuevaAPIEn(madrid), "/pistas/p-1/disponibilidad?fecha=2026-12-01")
	if len(horas) != 11 || horas[0] != "11:00" || horas[10] != "21:00" {
		t.Fatalf("esperaba de 11:00 a 21:00, salió %v", horas)
	}
}

func TestDisponibilidadConErrores(t *testing.T) {
	casos := []struct {
		nombre, ruta string
		estado       int
	}{
		{"sin fecha", "/pistas/p-1/disponibilidad", 400},
		{"fecha mal escrita", "/pistas/p-1/disponibilidad?fecha=02-12-2026", 400},
		{"pista que no existe", "/pistas/p-9/disponibilidad?fecha=2026-12-02", 404},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if estado, _, _ := horasLibres(t, nuevaAPI(), c.ruta); estado != c.estado {
				t.Fatalf("esperaba %d, salió %d", c.estado, estado)
			}
		})
	}
}
