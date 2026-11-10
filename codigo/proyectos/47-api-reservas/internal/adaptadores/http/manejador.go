// Package http traduce peticiones HTTP a llamadas a los casos de uso
// y sus resultados a respuestas JSON.
package http

import (
	"encoding/json"
	"net/http"
	"time"

	"reservas/internal/aplicacion"
)

// CasosDeUso agrupa lo que este adaptador sabe invocar.
type CasosDeUso struct {
	Reservar *aplicacion.ReservarPista
	Cancelar *aplicacion.CancelarReserva
	Listar   *aplicacion.ListarReservasDeSocio
	// Añadidos en el capítulo 47 para la web y la app (capítulos 49 y 50).
	Pistas         *aplicacion.ListarPistas
	Disponibilidad *aplicacion.ConsultarDisponibilidad
}

type Manejador struct {
	casos CasosDeUso
	zona  *time.Location // hora local del club, para leer y escribir fechas
}

func NuevoManejador(casos CasosDeUso, zona *time.Location) *Manejador {
	return &Manejador{casos: casos, zona: zona}
}

func (m *Manejador) Rutas() http.Handler {
	rutas := http.NewServeMux()
	rutas.HandleFunc("POST /reservas", m.reservar)
	rutas.HandleFunc("POST /reservas/{id}/cancelacion", m.cancelar)
	rutas.HandleFunc("GET /socios/{id}/reservas", m.listarReservas)
	rutas.HandleFunc("GET /pistas", m.listarPistas)
	rutas.HandleFunc("GET /pistas/{id}/disponibilidad", m.disponibilidad)
	return rutas
}

type reservaJSON struct {
	ID      string `json:"id"`
	SocioID string `json:"socio_id"`
	PistaID string `json:"pista_id"`
	Inicio  string `json:"inicio"`
	Fin     string `json:"fin"`
	Estado  string `json:"estado"`
}

func (m *Manejador) aJSON(r aplicacion.ReservaDTO) reservaJSON {
	return reservaJSON{
		ID:      r.ID,
		SocioID: r.SocioID,
		PistaID: r.PistaID,
		Inicio:  r.Inicio.In(m.zona).Format(time.RFC3339),
		Fin:     r.Fin.In(m.zona).Format(time.RFC3339),
		Estado:  r.Estado,
	}
}

// formatoDeEntrada es la hora local del club sin zona: "2026-12-02T10:00".
const formatoDeEntrada = "2006-01-02T15:04"

// La identidad ya no viaja en el cuerpo: la pone la capa de autenticación
// en la cabecera X-Socio-Id (ver identidad.go).
type peticionReserva struct {
	PistaID string `json:"pista_id"`
	Inicio  string `json:"inicio"`
}

func (m *Manejador) reservar(w http.ResponseWriter, r *http.Request) {
	socioID, ok := socioAutenticado(w, r)
	if !ok {
		return
	}
	var p peticionReserva
	if err := leerJSON(w, r, &p); err != nil || p.PistaID == "" {
		escribirJSON(w, http.StatusBadRequest, cuerpoDeError("peticion_no_valida",
			"se esperaba un JSON con pista_id e inicio"))
		return
	}
	inicio, err := time.ParseInLocation(formatoDeEntrada, p.Inicio, m.zona)
	if err != nil {
		escribirJSON(w, http.StatusBadRequest, cuerpoDeError("peticion_no_valida",
			"inicio tiene que tener el formato 2026-12-02T10:00"))
		return
	}
	reserva, err := m.casos.Reservar.Ejecutar(r.Context(), aplicacion.SolicitudReserva{
		SocioID: socioID, PistaID: p.PistaID, Inicio: inicio,
	})
	if err != nil {
		escribirError(w, err)
		return
	}
	escribirJSON(w, http.StatusCreated, m.aJSON(reserva))
}

func (m *Manejador) cancelar(w http.ResponseWriter, r *http.Request) {
	socioID, ok := socioAutenticado(w, r)
	if !ok {
		return
	}
	reserva, err := m.casos.Cancelar.Ejecutar(r.Context(), aplicacion.SolicitudCancelacion{
		ReservaID: r.PathValue("id"), SocioID: socioID,
	})
	if err != nil {
		escribirError(w, err)
		return
	}
	escribirJSON(w, http.StatusOK, m.aJSON(reserva))
}

func (m *Manejador) listarReservas(w http.ResponseWriter, r *http.Request) {
	socioID, ok := socioAutenticado(w, r)
	if !ok {
		return
	}
	// Cada socio solo ve sus reservas, aunque cambie el id de la ruta (capítulo 44).
	if r.PathValue("id") != socioID {
		escribirJSON(w, http.StatusForbidden, cuerpoDeError("prohibido", "solo puedes ver tus reservas"))
		return
	}
	reservas, err := m.casos.Listar.Ejecutar(r.Context(), socioID)
	if err != nil {
		escribirError(w, err)
		return
	}
	respuesta := make([]reservaJSON, 0, len(reservas))
	for _, reserva := range reservas {
		respuesta = append(respuesta, m.aJSON(reserva))
	}
	escribirJSON(w, http.StatusOK, respuesta)
}

type pistaJSON struct {
	ID      string `json:"id"`
	Nombre  string `json:"nombre"`
	Deporte string `json:"deporte"`
}

// Las pistas y sus horas libres son públicas: no hace falta X-Socio-Id.
func (m *Manejador) listarPistas(w http.ResponseWriter, r *http.Request) {
	pistas, err := m.casos.Pistas.Ejecutar(r.Context())
	if err != nil {
		escribirError(w, err)
		return
	}
	respuesta := make([]pistaJSON, 0, len(pistas))
	for _, p := range pistas {
		respuesta = append(respuesta, pistaJSON{ID: p.ID, Nombre: p.Nombre, Deporte: p.Deporte})
	}
	escribirJSON(w, http.StatusOK, respuesta)
}

type disponibilidadJSON struct {
	Fecha       string   `json:"fecha"`
	HorasLibres []string `json:"horas_libres"`
}

func (m *Manejador) disponibilidad(w http.ResponseWriter, r *http.Request) {
	dia, err := time.ParseInLocation(time.DateOnly, r.URL.Query().Get("fecha"), m.zona)
	if err != nil {
		escribirJSON(w, http.StatusBadRequest, cuerpoDeError("peticion_no_valida",
			"fecha tiene que tener el formato 2026-12-02"))
		return
	}
	disponibilidad, err := m.casos.Disponibilidad.Ejecutar(r.Context(), r.PathValue("id"), dia)
	if err != nil {
		escribirError(w, err)
		return
	}
	respuesta := disponibilidadJSON{
		Fecha:       disponibilidad.Fecha.Format(time.DateOnly),
		HorasLibres: make([]string, 0, len(disponibilidad.HorasLibres)),
	}
	for _, inicio := range disponibilidad.HorasLibres {
		respuesta.HorasLibres = append(respuesta.HorasLibres, inicio.In(m.zona).Format("15:04"))
	}
	escribirJSON(w, http.StatusOK, respuesta)
}

func leerJSON(w http.ResponseWriter, r *http.Request, destino any) error {
	decodificador := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16))
	decodificador.DisallowUnknownFields()
	return decodificador.Decode(destino)
}

func escribirJSON(w http.ResponseWriter, estado int, cuerpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	json.NewEncoder(w).Encode(cuerpo)
}
