// Package red tiene lo que comparten los adaptadores HTTP: reintentos y errores.
package red

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

// Politica dice cuántas veces y con qué espera se reintenta una petición.
type Politica struct {
	Intentos     int           // intentos en total, contando el primero
	EsperaBase   time.Duration // espera antes del segundo intento; luego se dobla
	EsperaMaxima time.Duration // tope de cada espera
}

// PoliticaInteractiva es para cuando hay una persona esperando delante de la terminal:
// pocos intentos y esperas cortas.
var PoliticaInteractiva = Politica{Intentos: 3, EsperaBase: time.Second, EsperaMaxima: 8 * time.Second}

// ErrorHTTP es una respuesta con un código de error que no se reintenta o que
// sigue fallando después de los reintentos.
type ErrorHTTP struct {
	Estado int
	Cuerpo string
}

func (e *ErrorHTTP) Error() string {
	return fmt.Sprintf("respuesta HTTP %d: %s", e.Estado, e.Cuerpo)
}

// Enviar hace la petición que construye crear y la reintenta si falla la red,
// si el servidor responde 429 (demasiadas peticiones) o un 5xx. Un 400, 401,
// 403 o 404 no se reintenta: repetirlo daría el mismo error.
func Enviar(ctx context.Context, cliente *http.Client, p Politica, crear func() (*http.Request, error)) ([]byte, error) {
	var ultimo error
	for intento := range p.Intentos {
		if intento > 0 {
			if err := dormir(ctx, p.espera(intento, ultimo)); err != nil {
				return nil, err
			}
		}
		req, err := crear()
		if err != nil {
			return nil, err
		}
		resp, err := cliente.Do(req.WithContext(ctx))
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			ultimo = err // fallo de red: se reintenta
			continue
		}
		cuerpo, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()
		if err != nil {
			ultimo = err
			continue
		}
		if resp.StatusCode < 300 {
			return cuerpo, nil
		}
		errHTTP := &ErrorHTTP{Estado: resp.StatusCode, Cuerpo: resumir(cuerpo)}
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			return nil, errHTTP
		}
		ultimo = &errorConEspera{ErrorHTTP: errHTTP, retryAfter: resp.Header.Get("Retry-After")}
	}
	if e, ok := ultimo.(*errorConEspera); ok {
		return nil, e.ErrorHTTP
	}
	return nil, fmt.Errorf("sin respuesta tras %d intentos: %w", p.Intentos, ultimo)
}

// errorConEspera guarda la cabecera Retry-After para respetarla en el siguiente intento.
type errorConEspera struct {
	*ErrorHTTP
	retryAfter string
}

// espera calcula un backoff exponencial con jitter (un componente aleatorio),
// para que muchos clientes que fallan a la vez no reintenten a la vez.
func (p Politica) espera(intento int, ultimo error) time.Duration {
	if e, ok := ultimo.(*errorConEspera); ok {
		if segundos, err := strconv.Atoi(e.retryAfter); err == nil {
			return min(time.Duration(segundos)*time.Second, p.EsperaMaxima)
		}
	}
	espera := min(p.EsperaBase<<(intento-1), p.EsperaMaxima)
	return espera/2 + rand.N(espera/2+1)
}

func dormir(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// resumir recorta el cuerpo de un error para que quepa en un mensaje.
func resumir(cuerpo []byte) string {
	const maximo = 300
	if len(cuerpo) > maximo {
		return string(cuerpo[:maximo]) + "..."
	}
	return string(cuerpo)
}
