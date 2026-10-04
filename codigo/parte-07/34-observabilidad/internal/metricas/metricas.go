// Package metricas guarda contadores y duraciones en memoria.
//
// Es una versión mínima para entender qué se mide. Un sistema real usaría
// histogramas con cubos fijos y una herramienta que los recoja, porque guardar
// todas las duraciones no escala.
package metricas

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

// Registro acumula peticiones por ruta y código de estado.
type Registro struct {
	mu         sync.Mutex
	contadores map[string]int
	duraciones map[string][]time.Duration
}

// NuevoRegistro crea un registro vacío.
func NuevoRegistro() *Registro {
	return &Registro{
		contadores: map[string]int{},
		duraciones: map[string][]time.Duration{},
	}
}

// Registrar apunta una petición terminada.
func (r *Registro) Registrar(ruta string, estado int, d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.contadores[fmt.Sprintf("%s %d", ruta, estado)]++
	r.duraciones[ruta] = append(r.duraciones[ruta], d)
}

// Contador devuelve cuántas peticiones a ruta acabaron con ese estado.
func (r *Registro) Contador(ruta string, estado int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.contadores[fmt.Sprintf("%s %d", ruta, estado)]
}

// Percentil devuelve la duración por debajo de la cual queda el p por ciento
// de las peticiones a ruta (método del rango más cercano).
func (r *Registro) Percentil(ruta string, p float64) time.Duration {
	r.mu.Lock()
	ds := slices.Clone(r.duraciones[ruta])
	r.mu.Unlock()
	return Percentil(ds, p)
}

// Percentil calcula el percentil p (entre 0 y 100) de una lista de duraciones.
func Percentil(ds []time.Duration, p float64) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	slices.Sort(ds)
	rango := int(math.Ceil(p / 100 * float64(len(ds))))
	if rango < 1 {
		rango = 1
	}
	return ds[rango-1]
}

// Resumen devuelve los contadores ordenados, una línea por ruta y estado.
func (r *Registro) Resumen() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	claves := make([]string, 0, len(r.contadores))
	for k := range r.contadores {
		claves = append(claves, k)
	}
	sort.Strings(claves)
	var b strings.Builder
	for _, k := range claves {
		fmt.Fprintf(&b, "%-22s %d\n", k, r.contadores[k])
	}
	return b.String()
}
