// Package acoplado es el contraejemplo del capítulo 16: una sola función
// que lee HTTP, aplica las reglas, guarda y avisa. Funciona, y por eso engaña.
package acoplado

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	mu       sync.Mutex
	ocupadas = map[string][]time.Time{} // pista -> horas reservadas
	porSocio = map[string]int{}         // socio -> número de reservas
)

func ManejarReserva(w http.ResponseWriter, r *http.Request) {
	var cuerpo struct {
		Socio  string `json:"socio"`
		Pista  string `json:"pista"`
		Inicio string `json:"inicio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		http.Error(w, "JSON no válido", http.StatusBadRequest)
		return
	}
	inicio, err := time.Parse("2006-01-02T15:04", cuerpo.Inicio)
	if err != nil {
		http.Error(w, "fecha no válida", http.StatusBadRequest)
		return
	}
	if inicio.Minute() != 0 || inicio.Hour() < 8 || inicio.Hour() > 21 {
		http.Error(w, "horario no válido", http.StatusUnprocessableEntity)
		return
	}
	if inicio.Before(time.Now()) || inicio.After(time.Now().AddDate(0, 0, 14)) {
		http.Error(w, "fuera de plazo", http.StatusUnprocessableEntity)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	if porSocio[cuerpo.Socio] >= 2 {
		http.Error(w, "demasiadas reservas", http.StatusUnprocessableEntity)
		return
	}
	for _, otra := range ocupadas[cuerpo.Pista] {
		if otra.Equal(inicio) {
			http.Error(w, "pista ocupada", http.StatusConflict)
			return
		}
	}
	ocupadas[cuerpo.Pista] = append(ocupadas[cuerpo.Pista], inicio)
	porSocio[cuerpo.Socio]++
	log.Printf("correo a %s: reserva confirmada", cuerpo.Socio)
	w.WriteHeader(http.StatusCreated)
}
