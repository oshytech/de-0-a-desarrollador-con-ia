package http

import (
	"errors"
	"log"
	"net/http"

	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

// estadosPorRegla dice qué código HTTP corresponde a cada regla incumplida.
// Las que no aparecen aquí se responden con 422.
var estadosPorRegla = map[*dominio.ReglaIncumplida]int{}

func escribirError(w http.ResponseWriter, err error) {
	var regla *dominio.ReglaIncumplida
	switch {
	case errors.Is(err, aplicacion.ErrNoEncontrado):
		escribirJSON(w, http.StatusNotFound, cuerpoDeError("no_encontrado", err.Error()))
	case errors.As(err, &regla):
		estado, ok := estadosPorRegla[regla]
		if !ok {
			estado = http.StatusUnprocessableEntity
		}
		escribirJSON(w, estado, cuerpoDeError(regla.Codigo, regla.Mensaje))
	default:
		// Un error que no esperábamos es un fallo nuestro: se registra entero
		// y al cliente no se le cuentan los detalles internos.
		log.Printf("error inesperado: %v", err)
		escribirJSON(w, http.StatusInternalServerError, cuerpoDeError("error_interno", "error interno"))
	}
}

func cuerpoDeError(codigo, mensaje string) map[string]string {
	return map[string]string{"error": codigo, "mensaje": mensaje}
}
