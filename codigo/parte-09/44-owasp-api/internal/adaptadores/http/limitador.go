package http

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Limitador implementa un cubo de fichas (token bucket) por clave. Cada
// clave empieza con Rafaga fichas, cada petición gasta una y se recupera
// una ficha cada Intervalo. Sin fichas, la petición se rechaza con 429.
type Limitador struct {
	Rafaga    int
	Intervalo time.Duration
	Ahora     func() time.Time
	Clave     func(*http.Request) string

	mu     sync.Mutex
	cubos  map[string]*cubo
	limpio time.Time
}

type cubo struct {
	fichas float64
	visto  time.Time
}

// Permitir devuelve si la petición pasa y, si no, cuánto hay que esperar.
func (l *Limitador) Permitir(clave string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	ahora := l.Ahora()
	if l.cubos == nil {
		l.cubos = map[string]*cubo{}
	}
	l.purgar(ahora)

	c, ok := l.cubos[clave]
	if !ok {
		c = &cubo{fichas: float64(l.Rafaga), visto: ahora}
		l.cubos[clave] = c
	}
	recuperadas := float64(ahora.Sub(c.visto)) / float64(l.Intervalo)
	c.fichas = math.Min(float64(l.Rafaga), c.fichas+recuperadas)
	c.visto = ahora

	if c.fichas >= 1 {
		c.fichas--
		return true, 0
	}
	falta := time.Duration((1 - c.fichas) * float64(l.Intervalo))
	return false, falta
}

// purgar borra los cubos llenos que nadie usa. Sin esto, un atacante que
// cambia de clave en cada petición haría crecer el mapa sin límite: el
// propio limitador sería un consumo de recursos sin restringir.
func (l *Limitador) purgar(ahora time.Time) {
	llenoTras := time.Duration(l.Rafaga) * l.Intervalo
	if ahora.Sub(l.limpio) < llenoTras {
		return
	}
	for k, c := range l.cubos {
		if ahora.Sub(c.visto) >= llenoTras {
			delete(l.cubos, k)
		}
	}
	l.limpio = ahora
}

func (l *Limitador) Middleware(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, espera := l.Permitir(l.Clave(r))
		if !ok {
			segundos := int(math.Ceil(espera.Seconds()))
			w.Header().Set("Retry-After", strconv.Itoa(segundos))
			responder(w, http.StatusTooManyRequests, map[string]string{"mensaje": "demasiadas peticiones, prueba más tarde"})
			return
		}
		siguiente.ServeHTTP(w, r)
	})
}

// PorSocio usa como clave el socio autenticado. Se aplica después de
// Identificar, así que el socio ya está en el contexto.
func PorSocio(r *http.Request) string { return "socio:" + strconv.FormatInt(socioDe(r.Context()), 10) }

func (l *Limitador) numeroDeCubos() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.cubos)
}
