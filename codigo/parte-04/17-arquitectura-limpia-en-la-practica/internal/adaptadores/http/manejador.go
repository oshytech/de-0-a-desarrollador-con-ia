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
	Listar *aplicacion.ListarReservasDeSocio
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

func escribirJSON(w http.ResponseWriter, estado int, cuerpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	json.NewEncoder(w).Encode(cuerpo)
}
