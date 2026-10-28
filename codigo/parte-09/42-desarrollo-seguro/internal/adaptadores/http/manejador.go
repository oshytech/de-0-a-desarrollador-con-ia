package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"desarrolloseguro/internal/dominio"
)

// ManejadorReservar valida la petición en dos pasos: la forma en el borde
// (DecodificarSolicitud) y el significado en el dominio (ValidarFranja).
type ManejadorReservar struct {
	Ahora    func() time.Time
	ZonaClub *time.Location
	Log      *slog.Logger
}

func (m ManejadorReservar) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	datos, err := DecodificarSolicitud(w, r)
	var errEntrada ErrEntrada
	switch {
	case errors.Is(err, ErrCuerpoDemasiadoGrande):
		responder(w, http.StatusRequestEntityTooLarge, "la petición es demasiado grande")
		return
	case errors.As(err, &errEntrada):
		responder(w, http.StatusBadRequest, errEntrada.Motivo)
		return
	case err != nil:
		m.Log.Error("error inesperado al leer la reserva", "error", err)
		responder(w, http.StatusInternalServerError, "error interno")
		return
	}

	if err := dominio.ValidarFranja(datos.Inicio, m.Ahora(), m.ZonaClub); err != nil {
		// Los errores de dominio hablan de la petición del propio usuario,
		// así que se pueden devolver tal cual: no revelan nada interno.
		responder(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// Aquí se llamaría al caso de uso ReservarPista. En este ejemplo basta
	// con confirmar que la entrada ha pasado los dos filtros.
	responder(w, http.StatusAccepted, "solicitud válida")
}

func responder(w http.ResponseWriter, estado int, mensaje string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(map[string]string{"mensaje": mensaje})
}
