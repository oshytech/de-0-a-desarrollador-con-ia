package red

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

var rapida = Politica{Intentos: 3, EsperaBase: time.Millisecond, EsperaMaxima: 5 * time.Millisecond}

// servidor responde con los códigos indicados, uno por petición.
func servidor(t *testing.T, codigos ...int) (*httptest.Server, *atomic.Int32) {
	var llamadas atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(llamadas.Add(1)) - 1
		w.WriteHeader(codigos[min(n, len(codigos)-1)])
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(s.Close)
	return s, &llamadas
}

func enviar(s *httptest.Server, ctx context.Context) ([]byte, error) {
	return Enviar(ctx, s.Client(), rapida, func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, s.URL, nil)
	})
}

func TestReintentaErroresTemporales(t *testing.T) {
	s, llamadas := servidor(t, 503, 429, 200)
	cuerpo, err := enviar(s, context.Background())
	if err != nil || string(cuerpo) != `{"ok":true}` || llamadas.Load() != 3 {
		t.Fatalf("cuerpo %q, error %v, llamadas %d", cuerpo, err, llamadas.Load())
	}
}

func TestNoReintentaErroresDelCliente(t *testing.T) {
	s, llamadas := servidor(t, 401)
	_, err := enviar(s, context.Background())
	var errHTTP *ErrorHTTP
	if !errors.As(err, &errHTTP) || errHTTP.Estado != 401 || llamadas.Load() != 1 {
		t.Fatalf("error %v, llamadas %d", err, llamadas.Load())
	}
}

func TestSeRindeTrasLosIntentos(t *testing.T) {
	s, llamadas := servidor(t, 500)
	_, err := enviar(s, context.Background())
	var errHTTP *ErrorHTTP
	if !errors.As(err, &errHTTP) || errHTTP.Estado != 500 || llamadas.Load() != 3 {
		t.Fatalf("error %v, llamadas %d", err, llamadas.Load())
	}
}

func TestRespetaLaCancelacion(t *testing.T) {
	s, _ := servidor(t, 500)
	ctx, cancelar := context.WithCancel(context.Background())
	cancelar()
	if _, err := enviar(s, ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, quiero context.Canceled", err)
	}
}

func TestEsperaConJitterDentroDelTope(t *testing.T) {
	p := Politica{Intentos: 5, EsperaBase: time.Second, EsperaMaxima: 8 * time.Second}
	for intento := 1; intento < 5; intento++ {
		e := p.espera(intento, nil)
		techo := min(time.Second<<(intento-1), 8*time.Second)
		if e < techo/2 || e > techo {
			t.Fatalf("intento %d: espera %v fuera de [%v, %v]", intento, e, techo/2, techo)
		}
	}
	conCabecera := &errorConEspera{ErrorHTTP: &ErrorHTTP{Estado: 429}, retryAfter: "2"}
	if e := p.espera(1, conCabecera); e != 2*time.Second {
		t.Fatalf("con Retry-After: 2 la espera es %v", e)
	}
}
