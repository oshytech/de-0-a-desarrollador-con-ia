// Package dominio: el socio del club.
package dominio

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

// Socio tiene campos que el propio socio puede cambiar (nombre, correo) y
// campos que solo cambia la administración del club (activo, notas).
type Socio struct {
	ID            int64
	Nombre        string
	Correo        string
	Activo        bool   // regla 4: solo los socios activos pueden reservar
	NotasInternas string // p. ej. "cuota de octubre pendiente"
}

var (
	ErrNombreNoValido = errors.New("el nombre tiene que tener entre 2 y 80 caracteres")
	ErrCorreoNoValido = errors.New("el correo no es válido")
)

func ValidarNombre(n string) error {
	n = strings.TrimSpace(n)
	if l := utf8.RuneCountInString(n); l < 2 || l > 80 {
		return ErrNombreNoValido
	}
	return nil
}

func ValidarCorreo(c string) error {
	dir, err := mail.ParseAddress(c)
	if err != nil || dir.Address != c || len(c) > 254 {
		return ErrCorreoNoValido
	}
	return nil
}
