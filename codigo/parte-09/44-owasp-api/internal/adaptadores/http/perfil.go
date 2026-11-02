// Package http: el perfil del socio (API3, autorización a nivel de propiedad)
// y la limitación de peticiones (API4).
package http

import (
	"encoding/json"
	"net/http"
	"sync"

	"owaspapi/internal/dominio"
)

// Socios es un almacén mínimo en memoria para el ejemplo.
type Socios struct {
	mu    sync.Mutex
	Datos map[int64]dominio.Socio
}

func (s *Socios) leer(id int64) dominio.Socio {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Datos[id]
}

func (s *Socios) guardar(x dominio.Socio) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Datos[x.ID] = x
}

// socioCompleto refleja el modelo interno tal cual. Usarlo para entrada o
// salida de la API es el origen de los dos fallos de API3.
type socioCompleto struct {
	ID            int64  `json:"id"`
	Nombre        string `json:"nombre"`
	Correo        string `json:"correo"`
	Activo        bool   `json:"activo"`
	NotasInternas string `json:"notas_internas"`
}

// ActualizarPerfilVulnerable copia al socio todo lo que llega en el cuerpo
// (mass assignment) y devuelve el objeto entero (exposición excesiva).
func ActualizarPerfilVulnerable(socios *Socios) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := socioDe(r.Context())
		actual := socios.leer(id)
		c := socioCompleto(actual)
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			responder(w, http.StatusBadRequest, map[string]string{"mensaje": "JSON no válido"})
			return
		}
		c.ID = id
		socios.guardar(dominio.Socio(c))
		responder(w, http.StatusOK, c)
	}
}

// CambiosPerfil es el contrato de entrada: solo lo que el socio puede tocar.
// Los punteros distinguen "no lo envía" de "lo envía vacío".
type CambiosPerfil struct {
	Nombre *string `json:"nombre"`
	Correo *string `json:"correo"`
}

// PerfilRespuesta es el contrato de salida: lo que el socio necesita ver.
type PerfilRespuesta struct {
	Nombre string `json:"nombre"`
	Correo string `json:"correo"`
	Activo bool   `json:"activo"`
}

func ActualizarPerfil(socios *Socios) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields() // "activo" o "notas_internas" en el cuerpo: 400

		var cambios CambiosPerfil
		if err := dec.Decode(&cambios); err != nil {
			responder(w, http.StatusBadRequest, map[string]string{"mensaje": "solo se pueden cambiar nombre y correo"})
			return
		}

		s := socios.leer(socioDe(r.Context()))
		if cambios.Nombre != nil {
			if err := dominio.ValidarNombre(*cambios.Nombre); err != nil {
				responder(w, http.StatusBadRequest, map[string]string{"mensaje": err.Error()})
				return
			}
			s.Nombre = *cambios.Nombre
		}
		if cambios.Correo != nil {
			if err := dominio.ValidarCorreo(*cambios.Correo); err != nil {
				responder(w, http.StatusBadRequest, map[string]string{"mensaje": err.Error()})
				return
			}
			s.Correo = *cambios.Correo
		}
		socios.guardar(s)
		responder(w, http.StatusOK, PerfilRespuesta{Nombre: s.Nombre, Correo: s.Correo, Activo: s.Activo})
	}
}

func responder(w http.ResponseWriter, estado int, cuerpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(cuerpo)
}
