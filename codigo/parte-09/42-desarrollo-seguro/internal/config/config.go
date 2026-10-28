// Package config carga la configuración del servicio desde el entorno.
// Falla al arrancar si falta algo y nunca imprime el valor de un secreto.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

// Secreto es un string que no se deja imprimir. Si alguien hace
// fmt.Println(cfg) o lo mete en un log, sale "[oculto]".
type Secreto string

func (Secreto) String() string                 { return "[oculto]" }
func (Secreto) GoString() string               { return "[oculto]" }
func (Secreto) LogValue() slog.Value           { return slog.StringValue("[oculto]") }
func (s Secreto) MarshalText() ([]byte, error) { return []byte("[oculto]"), nil }

// Revelar devuelve el valor real. Tener que llamarlo a propósito hace que
// cada uso del secreto sea fácil de encontrar en una revisión.
func (s Secreto) Revelar() string { return string(s) }

// Config agrupa lo que el servicio necesita. Los campos son exportados para
// que fmt pueda llamar a los métodos de Secreto (con campos no exportados,
// fmt imprime el valor crudo).
type Config struct {
	Entorno      string
	Direccion    string
	URLBaseDatos Secreto
	ClaveTokens  Secreto
}

const longitudMinimaClave = 32

// Cargar lee la configuración con la función que se le pase (os.Getenv en
// producción, un mapa en los tests). Devuelve todos los problemas juntos.
func Cargar(leer func(string) string) (Config, error) {
	c := Config{
		Entorno: valorODefecto(leer("RESERVAS_ENTORNO"), "desarrollo"),
		// Por defecto solo escucha en la propia máquina: exponerlo es una decisión.
		Direccion:    valorODefecto(leer("RESERVAS_DIRECCION"), "127.0.0.1:8080"),
		URLBaseDatos: Secreto(leer("RESERVAS_URL_BD")),
		ClaveTokens:  Secreto(leer("RESERVAS_CLAVE_TOKENS")),
	}

	var problemas []string
	if c.URLBaseDatos == "" {
		problemas = append(problemas, "falta RESERVAS_URL_BD")
	}
	if c.ClaveTokens == "" {
		problemas = append(problemas, "falta RESERVAS_CLAVE_TOKENS")
	} else if len(c.ClaveTokens) < longitudMinimaClave {
		// Se dice qué está mal, nunca cuál es el valor.
		problemas = append(problemas, fmt.Sprintf("RESERVAS_CLAVE_TOKENS tiene que tener al menos %d caracteres", longitudMinimaClave))
	}
	if c.Entorno != "desarrollo" && c.Entorno != "pruebas" && c.Entorno != "produccion" {
		problemas = append(problemas, "RESERVAS_ENTORNO tiene que ser desarrollo, pruebas o produccion")
	}
	if len(problemas) > 0 {
		return Config{}, errors.New("configuración no válida: " + strings.Join(problemas, "; "))
	}
	return c, nil
}

func valorODefecto(v, defecto string) string {
	if v == "" {
		return defecto
	}
	return v
}
