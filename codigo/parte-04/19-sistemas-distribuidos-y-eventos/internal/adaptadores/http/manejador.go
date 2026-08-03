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

type peticionReserva struct {
	SocioID string `json:"socio_id"`
	PistaID string `json:"pista_id"`
	Inicio  string `json:"inicio"`
}

func (m *Manejador) reservar(w http.ResponseWriter, r *http.Request) {
	var p peticionReserva
	if err := leerJSON(w, r, &p); err != nil || p.SocioID == "" || p.PistaID == "" {
		escribirJSON(w, http.StatusBadRequest, cuerpoDeError("peticion_no_valida",
			"se esperaba un JSON con socio_id, pista_id e inicio"))
		return
	}
	inicio, err := time.ParseInLocation(formatoDeEntrada, p.Inicio, m.zona)
	if err != nil {
		escribirJSON(w, http.StatusBadRequest, cuerpoDeError("peticion_no_valida",
			"inicio tiene que tener el formato 2026-12-02T10:00"))
		return
	}
	reserva, err := m.casos.Reservar.Ejecutar(r.Context(), aplicacion.SolicitudReserva{
		SocioID: p.SocioID, PistaID: p.PistaID, Inicio: inicio,
	})
	if err != nil {
		escribirError(w, err)
		return
	}
	escribirJSON(w, http.StatusCreated, m.aJSON(reserva))
}

type peticionCancelacion struct {
	SocioID string `json:"socio_id"`
}

func (m *Manejador) cancelar(w http.ResponseWriter, r *http.Request) {
	var p peticionCancelacion
	if err := leerJSON(w, r, &p); err != nil || p.SocioID == "" {
		escribirJSON(w, http.StatusBadRequest, cuerpoDeError("peticion_no_valida", "se esperaba un JSON con socio_id"))
		return
	}
	reserva, err := m.casos.Cancelar.Ejecutar(r.Context(), aplicacion.SolicitudCancelacion{
		ReservaID: r.PathValue("id"), SocioID: p.SocioID,
	})
	if err != nil {
		escribirError(w, err)
		return
	}
	escribirJSON(w, http.StatusOK, m.aJSON(reserva))
}

func (m *Manejador) listarReservas(w http.ResponseWriter, r *http.Request) {
	reservas, err := m.casos.Listar.Ejecutar(r.Context(), r.PathValue("id"))
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
