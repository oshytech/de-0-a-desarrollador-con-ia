// Package http: identificación del socio, la ruta de cancelación y una
// única forma de responder a los errores.
package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"owaspweb/internal/dominio"
)

type claveContexto struct{}

// Sesiones traduce un token opaco a un socio. En un sistema real sería una
// sesión en servidor o un token firmado que se valida; lo importante es que
// la identidad sale de aquí y nunca de un parámetro que manda el cliente.
type Sesiones map[string]int64

func (s Sesiones) Identificar(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		socio, existe := s[token]
		if !ok || !existe {
			responder(w, http.StatusUnauthorized, "no autenticado")
			return
		}
		ctx := context.WithValue(r.Context(), claveContexto{}, socio)
		siguiente.ServeHTTP(w, r.WithContext(ctx))
	})
}

func socioDe(ctx context.Context) int64 { return ctx.Value(claveContexto{}).(int64) }

// Canceladora es lo que necesita el handler: el caso de uso seguro o el
// vulnerable cumplen esta interfaz.
type Canceladora interface {
	Ejecutar(ctx context.Context, socioID, reservaID int64) error
}

func CancelarReserva(c Canceladora, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id <= 0 {
			responder(w, http.StatusBadRequest, "identificador de reserva no válido")
			return
		}
		if err := c.Ejecutar(r.Context(), socioDe(r.Context()), id); err != nil {
			responderError(w, log, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// responderError traduce errores conocidos a respuestas claras y cualquier
// otro a un 500 genérico. El detalle se queda en el log con un identificador
// que el usuario puede dar a soporte (A10: fallar de forma segura).
func responderError(w http.ResponseWriter, log *slog.Logger, err error) {
	switch {
	case errors.Is(err, dominio.ErrReservaNoEncontrada):
		responder(w, http.StatusNotFound, err.Error())
	case errors.Is(err, dominio.ErrFueraDePlazo), errors.Is(err, dominio.ErrYaCancelada):
		responder(w, http.StatusConflict, err.Error())
	default:
		incidencia := nuevoID()
		log.Error("error no controlado", "incidencia", incidencia, "error", err)
		responder(w, http.StatusInternalServerError, "error interno, referencia "+incidencia)
	}
}

func responder(w http.ResponseWriter, estado int, mensaje string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(map[string]string{"mensaje": mensaje})
}

func nuevoID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
