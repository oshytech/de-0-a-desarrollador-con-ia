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
var estadosPorRegla = map[*dominio.ReglaIncumplida]int{
	dominio.ErrPistaOcupada:       http.StatusConflict,
	dominio.ErrReservaYaCancelada: http.StatusConflict,
}

func escribirError(w http.ResponseWriter, err error) {
	var regla *dominio.ReglaIncumplida
	switch {
	// La reserva de otro socio se responde igual que una que no existe:
	// un 403 confirmaría que ese id es válido (capítulo 44).
	case errors.Is(err, aplicacion.ErrNoEncontrado), errors.Is(err, dominio.ErrReservaDeOtroSocio):
		escribirJSON(w, http.StatusNotFound, cuerpoDeError("no_encontrado", "no existe"))
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
