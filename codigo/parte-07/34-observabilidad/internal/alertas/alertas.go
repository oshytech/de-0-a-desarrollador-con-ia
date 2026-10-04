// Package alertas decide cuándo avisar a una persona y cuándo callarse.
package alertas

import (
	"fmt"
	"sync"
	"time"
)

// Alerta es un aviso listo para enviar. Clave identifica el problema: dos
// alertas con la misma clave hablan de lo mismo.
type Alerta struct {
	Clave   string
	Mensaje string
}

// RevisarDuracion compara lo que ha tardado una tarea con su duración de
// referencia y devuelve una alerta si se ha pasado del factor permitido.
func RevisarDuracion(tarea string, duracion, referencia time.Duration, factor float64) (Alerta, bool) {
	limite := time.Duration(float64(referencia) * factor)
	if duracion <= limite {
		return Alerta{}, false
	}
	return Alerta{
		Clave: "duracion:" + tarea,
		Mensaje: fmt.Sprintf("%s ha tardado %s (referencia %s, límite %s)",
			tarea, duracion, referencia, limite),
	}, true
}

// Agrupador evita mandar la misma alerta una y otra vez. Dentro de la ventana
// solo se envía la primera de cada clave y las demás se cuentan.
type Agrupador struct {
	ventana     time.Duration
	mu          sync.Mutex
	ultimoEnvio map[string]time.Time
	suprimidas  map[string]int
}

// NuevoAgrupador crea un agrupador con la ventana indicada.
func NuevoAgrupador(ventana time.Duration) *Agrupador {
	return &Agrupador{
		ventana:     ventana,
		ultimoEnvio: map[string]time.Time{},
		suprimidas:  map[string]int{},
	}
}

// Decidir indica si hay que enviar la alerta ahora y, si es así, cuántas
// iguales se callaron desde el último envío.
func (g *Agrupador) Decidir(a Alerta, ahora time.Time) (enviar bool, calladas int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	ultimo, visto := g.ultimoEnvio[a.Clave]
	if visto && ahora.Sub(ultimo) < g.ventana {
		g.suprimidas[a.Clave]++
		return false, 0
	}
	calladas = g.suprimidas[a.Clave]
	g.suprimidas[a.Clave] = 0
	g.ultimoEnvio[a.Clave] = ahora
	return true, calladas
}
