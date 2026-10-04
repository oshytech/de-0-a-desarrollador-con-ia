// Package http contiene el middleware de observabilidad y los handlers del
// ejemplo, con net/http de la librería estándar.
package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"reservas/internal/metricas"
)

// CabeceraID es la cabecera por la que viaja el identificador de petición.
const CabeceraID = "X-Request-ID"

type claveLogger struct{}

// LoggerDe devuelve el logger de la petición, que ya lleva su identificador.
// Si el contexto no tiene ninguno, devuelve el logger por defecto.
func LoggerDe(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(claveLogger{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// Observar envuelve un handler. A cada petición le asigna un identificador
// (o reutiliza el que llega en la cabecera), lo devuelve en la respuesta,
// deja en el contexto un logger que lo incluye y, al terminar, escribe una
// línea de log y registra la duración en las métricas.
func Observar(base *slog.Logger, m *metricas.Registro, siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		id := r.Header.Get(CabeceraID)
		if id == "" || len(id) > 64 {
			id = nuevoID()
		}
		w.Header().Set(CabeceraID, id)

		log := base.With("request_id", id)
		r2 := r.WithContext(context.WithValue(r.Context(), claveLogger{}, log))
		rec := &grabadora{ResponseWriter: w, estado: http.StatusOK}

		siguiente.ServeHTTP(rec, r2)

		// r2.Pattern lo rellena el ServeMux: "POST /reservas", no la URL
		// concreta. Así las métricas no crean una serie por cada id.
		ruta := r2.Pattern
		if ruta == "" {
			ruta = "sin ruta"
		}
		duracion := time.Since(inicio)
		m.Registrar(ruta, rec.estado, duracion)

		nivel := slog.LevelInfo
		if rec.estado >= 500 {
			nivel = slog.LevelError
		}
		log.Log(r.Context(), nivel, "petición atendida",
			"ruta", ruta,
			"estado", rec.estado,
			"duracion_ms", float64(duracion.Microseconds())/1000,
		)
	})
}

type grabadora struct {
	http.ResponseWriter
	estado int
}

func (g *grabadora) WriteHeader(estado int) {
	g.estado = estado
	g.ResponseWriter.WriteHeader(estado)
}

func nuevoID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
