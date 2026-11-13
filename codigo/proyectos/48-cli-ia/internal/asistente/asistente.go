// Package asistente contiene la lógica del asistente: buscar, preguntar al
// modelo y recordar la conversación. No sabe qué proveedor hay detrás.
package asistente

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Rol de quien escribe un mensaje.
type Rol string

const (
	Usuario   Rol = "user"
	Asistente Rol = "assistant"
)

// Mensaje es un turno de la conversación.
type Mensaje struct {
	Rol   Rol
	Texto string
}

// Fuente es un resultado de la búsqueda web.
type Fuente struct {
	Titulo    string
	URL       string
	Contenido string
}

// Modelo es el puerto hacia el modelo de lenguaje. Cada proveedor
// (Anthropic, OpenAI, Ollama, el simulado) es un adaptador.
type Modelo interface {
	Completar(ctx context.Context, sistema string, mensajes []Mensaje) (string, error)
}

// Buscador es el puerto hacia el motor de búsqueda web.
type Buscador interface {
	Buscar(ctx context.Context, consulta string, maximo int) ([]Fuente, error)
}

// Respuesta es lo que devuelve el asistente a la interfaz.
type Respuesta struct {
	Texto   string
	Fuentes []Fuente
	Aviso   string // algo que el usuario debe saber, como que la búsqueda falló
}

// Opciones de comportamiento.
const (
	MaxFuentes        = 5
	MaxTurnosRecuerda = 6    // pares pregunta-respuesta que se reenvían al modelo
	MaxCaracteresPreg = 2000 // una pregunta más larga se rechaza
	MaxCaracteresFte  = 1500 // cada fuente se recorta a este tamaño
)

var ErrPreguntaVacia = errors.New("la pregunta está vacía")
var ErrPreguntaLarga = fmt.Errorf("la pregunta supera los %d caracteres", MaxCaracteresPreg)

// Conversacion es una sesión con memoria de los turnos anteriores.
type Conversacion struct {
	modelo    Modelo
	buscador  Buscador
	historial []Mensaje
}

func NuevaConversacion(modelo Modelo, buscador Buscador) *Conversacion {
	return &Conversacion{modelo: modelo, buscador: buscador}
}

// Olvidar borra la memoria de la conversación.
func (c *Conversacion) Olvidar() { c.historial = nil }

// Responder busca en la web, pide la respuesta al modelo y la recuerda.
func (c *Conversacion) Responder(ctx context.Context, pregunta string) (Respuesta, error) {
	pregunta = strings.TrimSpace(pregunta)
	if pregunta == "" {
		return Respuesta{}, ErrPreguntaVacia
	}
	if len([]rune(pregunta)) > MaxCaracteresPreg {
		return Respuesta{}, ErrPreguntaLarga
	}

	var respuesta Respuesta
	fuentes, err := c.buscador.Buscar(ctx, pregunta, MaxFuentes)
	if err != nil {
		// Si la búsqueda falla seguimos sin fuentes, pero lo decimos.
		if ctx.Err() != nil {
			return Respuesta{}, ctx.Err()
		}
		respuesta.Aviso = "la búsqueda web ha fallado, la respuesta no usa fuentes: " + err.Error()
		fuentes = nil
	}
	respuesta.Fuentes = fuentes

	mensajes := append(c.recientes(), Mensaje{Rol: Usuario, Texto: MensajeConFuentes(pregunta, fuentes)})
	texto, err := c.modelo.Completar(ctx, InstruccionesDelSistema, mensajes)
	if err != nil {
		return Respuesta{}, fmt.Errorf("el modelo no ha respondido: %w", err)
	}
	respuesta.Texto = strings.TrimSpace(texto)

	// En la memoria guardamos la pregunta sin las fuentes: ocupan mucho y
	// en el turno siguiente se buscan otras.
	c.historial = append(c.historial,
		Mensaje{Rol: Usuario, Texto: pregunta},
		Mensaje{Rol: Asistente, Texto: respuesta.Texto})
	return respuesta, nil
}

// recientes devuelve los últimos turnos de la conversación.
func (c *Conversacion) recientes() []Mensaje {
	desde := max(0, len(c.historial)-2*MaxTurnosRecuerda)
	return append([]Mensaje(nil), c.historial[desde:]...)
}
