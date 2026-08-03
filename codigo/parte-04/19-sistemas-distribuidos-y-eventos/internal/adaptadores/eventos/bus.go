// Package eventos implementa el puerto aplicacion.PublicadorEventos con un
// bus en memoria hecho con canales. Todo vive en el mismo proceso: si el
// proceso se para, los eventos que estaban en los canales se pierden.
package eventos

import (
	"context"
	"errors"
	"sync"
	"time"

	"reservas/internal/aplicacion"
	"reservas/internal/dominio"
)

var ErrBusCerrado = errors.New("el bus de eventos está cerrado")

// Manejador es la función que reacciona a un evento.
type Manejador func(ctx context.Context, e dominio.Evento) error

// Fallido es un evento que un suscriptor no pudo procesar tras los reintentos.
// Es la versión en memoria de una cola de mensajes fallidos.
type Fallido struct {
	Suscriptor string
	Evento     dominio.Evento
	Error      error
}

type suscripcion struct {
	nombre  string
	buzon   chan dominio.Evento
	manejar Manejador
}

type Bus struct {
	mu            sync.RWMutex
	suscripciones []*suscripcion
	cerrado       bool
	trabajando    sync.WaitGroup

	intentos int
	espera   time.Duration

	muFallidos sync.Mutex
	fallidos   []Fallido
}

var _ aplicacion.PublicadorEventos = (*Bus)(nil)

// NuevoBus crea un bus que intenta entregar cada evento hasta `intentos`
// veces, esperando un poco más tras cada fallo.
func NuevoBus(intentos int, espera time.Duration) *Bus {
	return &Bus{intentos: intentos, espera: espera}
}

// Suscribir arranca una goroutine que procesa, en orden, los eventos
// que le llegan a este suscriptor.
func (b *Bus) Suscribir(nombre string, manejar Manejador) {
	s := &suscripcion{nombre: nombre, buzon: make(chan dominio.Evento, 100), manejar: manejar}
	b.mu.Lock()
	b.suscripciones = append(b.suscripciones, s)
	b.mu.Unlock()

	b.trabajando.Add(1)
	go func() {
		defer b.trabajando.Done()
		for evento := range s.buzon {
			b.entregar(s, evento)
		}
	}()
}

// Publicar deja una copia de cada evento en el buzón de cada suscriptor
// y vuelve sin esperar a que lo procesen.
func (b *Bus) Publicar(ctx context.Context, eventos ...dominio.Evento) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.cerrado {
		return ErrBusCerrado
	}
	for _, evento := range eventos {
		for _, s := range b.suscripciones {
			select {
			case s.buzon <- evento:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

func (b *Bus) entregar(s *suscripcion, evento dominio.Evento) {
	var err error
	for intento := 1; intento <= b.intentos; intento++ {
		if err = s.manejar(context.Background(), evento); err == nil {
			return
		}
		time.Sleep(time.Duration(intento) * b.espera)
	}
	b.muFallidos.Lock()
	b.fallidos = append(b.fallidos, Fallido{Suscriptor: s.nombre, Evento: evento, Error: err})
	b.muFallidos.Unlock()
}

// Fallidos devuelve los eventos que no se pudieron procesar.
func (b *Bus) Fallidos() []Fallido {
	b.muFallidos.Lock()
	defer b.muFallidos.Unlock()
	return append([]Fallido(nil), b.fallidos...)
}

// Cerrar deja de aceptar eventos y espera a que se procesen los pendientes.
func (b *Bus) Cerrar() {
	b.mu.Lock()
	if !b.cerrado {
		b.cerrado = true
		for _, s := range b.suscripciones {
			close(s.buzon)
		}
	}
	b.mu.Unlock()
	b.trabajando.Wait()
}
