package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"reservas/internal/dominio"
)

// Reservar es el caso de uso que el handler necesita. En el proyecto completo
// lo implementa aplicacion.ReservarPista.
type Reservar func(ctx context.Context, socioID, pistaID string, inicio time.Time) error

type peticionReserva struct {
	Socio  string    `json:"socio"`
	Pista  string    `json:"pista"`
	Inicio time.Time `json:"inicio"`
}

// ManejadorReservar atiende POST /reservas.
func ManejadorReservar(reservar Reservar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := LoggerDe(r.Context())
		var p peticionReserva
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "cuerpo no válido", http.StatusBadRequest)
			return
		}
		err := reservar(r.Context(), p.Socio, p.Pista, p.Inicio)
		if err == nil {
			w.WriteHeader(http.StatusCreated)
			return
		}
		if estado, esDeNegocio := estadoPara(err); esDeNegocio {
			// Un rechazo de negocio no es un fallo del sistema: se apunta
			// para poder contarlo, pero no despierta a nadie.
			log.Info("reserva rechazada", "motivo", err.Error(), "pista", p.Pista)
			http.Error(w, err.Error(), estado)
			return
		}
		log.Error("no se pudo reservar", "error", err, "socio", p.Socio, "pista", p.Pista)
		http.Error(w, "error interno", http.StatusInternalServerError)
	}
}

func estadoPara(err error) (int, bool) {
	switch {
	case errors.Is(err, dominio.ErrPistaOcupada), errors.Is(err, dominio.ErrLimiteReservas):
		return http.StatusConflict, true
	case errors.Is(err, dominio.ErrHorarioNoValido), errors.Is(err, dominio.ErrFueraDePlazo):
		return http.StatusUnprocessableEntity, true
	case errors.Is(err, dominio.ErrSocioInactivo):
		return http.StatusForbidden, true
	}
	return 0, false
}

// Comprobacion verifica una dependencia (base de datos, otro servicio).
type Comprobacion func(ctx context.Context) error

// ManejadorSalud atiende GET /health. Responde 200 si todas las dependencias
// contestan y 503 si alguna falla, con el detalle de cada una.
func ManejadorSalud(comprobaciones map[string]Comprobacion) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelar := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancelar()
		estado := http.StatusOK
		detalle := map[string]string{}
		for nombre, comprobar := range comprobaciones {
			if err := comprobar(ctx); err != nil {
				estado = http.StatusServiceUnavailable
				detalle[nombre] = err.Error()
				continue
			}
			detalle[nombre] = "ok"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(estado)
		_ = json.NewEncoder(w).Encode(detalle)
	}
}
