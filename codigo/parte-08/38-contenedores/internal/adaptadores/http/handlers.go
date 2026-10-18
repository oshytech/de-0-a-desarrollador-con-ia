// Package http expone la API de reservas con la librería estándar.
// En el capítulo 47, entre estos handlers y el repositorio está el caso
// de uso ReservarPista. Aquí se omite para centrarnos en el contenedor.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"reservas/internal/dominio"
)

// Reservador es lo único que estos handlers necesitan del almacenamiento.
type Reservador interface {
	Reservar(ctx context.Context, socioID, pistaID int64, inicio time.Time) (int64, error)
}

// Comprobador dice si las dependencias (la base de datos) responden.
type Comprobador interface {
	Ping(ctx context.Context) error
}

func NuevoMux(r Reservador, bd Comprobador) *http.ServeMux {
	mux := http.NewServeMux()

	// Vivo: el proceso responde. No toca la base de datos a propósito:
	// si la base de datos cae, reiniciar la API no la arregla.
	mux.HandleFunc("GET /vivo", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	// Listo: puede atender peticiones, porque la base de datos responde.
	mux.HandleFunc("GET /listo", func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()
		if err := bd.Ping(ctx); err != nil {
			http.Error(w, "base de datos no disponible", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /reservas", func(w http.ResponseWriter, req *http.Request) {
		var p struct {
			SocioID int64     `json:"socio_id"`
			PistaID int64     `json:"pista_id"`
			Inicio  time.Time `json:"inicio"`
		}
		if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
			http.Error(w, "JSON no válido", http.StatusBadRequest)
			return
		}
		id, err := r.Reservar(req.Context(), p.SocioID, p.PistaID, p.Inicio)
		switch {
		case err == nil:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]int64{"id": id})
		case errors.Is(err, dominio.ErrPistaOcupada):
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.Is(err, dominio.ErrSocioNoExiste):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, dominio.ErrHorarioNoValido):
			// El error lleva dentro el detalle de PostgreSQL: no se lo enseñamos al cliente.
			http.Error(w, dominio.ErrHorarioNoValido.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, dominio.ErrSocioInactivo), errors.Is(err, dominio.ErrLimiteReservas):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		default:
			slog.Error("reservar", "error", err)
			http.Error(w, "error interno", http.StatusInternalServerError)
		}
	})
	return mux
}
