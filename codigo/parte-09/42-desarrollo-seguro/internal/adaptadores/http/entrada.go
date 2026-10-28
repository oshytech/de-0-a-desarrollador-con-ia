// Package http contiene el borde HTTP: aquí se comprueba la forma de los datos
// que llegan de fuera antes de que el dominio los vea.
package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"regexp"
	"time"
)

// TamanoMaximoCuerpo es el tamaño máximo de una petición de reserva. Una
// reserva válida ocupa menos de cien bytes, así que 1 KiB sobra.
const TamanoMaximoCuerpo = 1 << 10

// Una pista se identifica como "tenis-3" o "padel-12". Lista de lo permitido
// (allowlist), no de lo prohibido.
var formatoPista = regexp.MustCompile(`^(tenis|padel)-[1-9][0-9]?$`)

// SolicitudReserva es lo único que el cliente puede enviar. El socio no
// viaja en el cuerpo: sale de la sesión.
type SolicitudReserva struct {
	PistaID string `json:"pista_id"`
	Inicio  string `json:"inicio"`
}

// DatosReserva es la solicitud ya validada y convertida a tipos de Go.
type DatosReserva struct {
	PistaID string
	Inicio  time.Time
}

// ErrEntrada es un error de forma: la petición no tiene el aspecto esperado.
type ErrEntrada struct{ Motivo string }

func (e ErrEntrada) Error() string { return e.Motivo }

var ErrCuerpoDemasiadoGrande = errors.New("cuerpo demasiado grande")

// DecodificarSolicitud lee y valida la forma de una solicitud de reserva.
func DecodificarSolicitud(w http.ResponseWriter, r *http.Request) (DatosReserva, error) {
	tipo, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || tipo != "application/json" {
		return DatosReserva{}, ErrEntrada{"el cuerpo tiene que ser JSON"}
	}

	r.Body = http.MaxBytesReader(w, r.Body, TamanoMaximoCuerpo)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // un campo que no esperamos es un error, no se ignora

	var s SolicitudReserva
	if err := dec.Decode(&s); err != nil {
		var demasiado *http.MaxBytesError
		if errors.As(err, &demasiado) {
			return DatosReserva{}, ErrCuerpoDemasiadoGrande
		}
		return DatosReserva{}, ErrEntrada{"JSON no válido o con campos no permitidos"}
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return DatosReserva{}, ErrEntrada{"el cuerpo tiene que contener un solo objeto JSON"}
	}

	if !formatoPista.MatchString(s.PistaID) {
		return DatosReserva{}, ErrEntrada{"pista_id no tiene un formato válido"}
	}
	inicio, err := time.Parse(time.RFC3339, s.Inicio)
	if err != nil {
		return DatosReserva{}, ErrEntrada{fmt.Sprintf("inicio tiene que ser una fecha RFC 3339, por ejemplo %q", "2026-12-01T18:00:00+01:00")}
	}
	return DatosReserva{PistaID: s.PistaID, Inicio: inicio}, nil
}
