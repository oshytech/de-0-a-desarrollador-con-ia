package http

import (
	"context"
	"net/http"
	"strings"
)

type claveContexto struct{}

// Sesiones traduce un token opaco a un socio (ver el capítulo 43). La
// identidad sale de aquí, nunca de un campo del cuerpo ni de la URL.
type Sesiones map[string]int64

func (s Sesiones) Identificar(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		socio, existe := s[token]
		if !ok || !existe {
			responder(w, http.StatusUnauthorized, map[string]string{"mensaje": "no autenticado"})
			return
		}
		siguiente.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claveContexto{}, socio)))
	})
}

func socioDe(ctx context.Context) int64 { return ctx.Value(claveContexto{}).(int64) }
