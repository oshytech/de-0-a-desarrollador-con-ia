// Demo del capítulo 34: levanta el servidor en memoria, le hace peticiones y
// enseña lo que queda registrado (logs JSON, métricas y alertas).
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	adaptador "reservas/internal/adaptadores/http"
	"reservas/internal/alertas"
	"reservas/internal/dominio"
	"reservas/internal/metricas"
)

// reservarFalso imita el caso de uso: la pista-2 siempre está ocupada, la
// pista-9 simula una base de datos caída y una de cada veinte peticiones
// tarda mucho más que las demás.
func reservarFalso() adaptador.Reservar {
	n := 0
	return func(ctx context.Context, socioID, pistaID string, inicio time.Time) error {
		n++
		if n%20 == 0 {
			time.Sleep(80 * time.Millisecond)
		} else {
			time.Sleep(2 * time.Millisecond)
		}
		switch pistaID {
		case "pista-2":
			return dominio.ErrPistaOcupada
		case "pista-9":
			return fmt.Errorf("guardar reserva: %w", errors.New("conexión rechazada"))
		}
		return nil
	}
}

func nuevoServidor(log *slog.Logger, m *metricas.Registro) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("POST /reservas", adaptador.ManejadorReservar(reservarFalso()))
	mux.Handle("GET /health", adaptador.ManejadorSalud(map[string]adaptador.Comprobacion{
		"base_de_datos": func(context.Context) error { return nil },
	}))
	return httptest.NewServer(adaptador.Observar(log, m, mux))
}

func reservar(url, pista string) {
	cuerpo := fmt.Sprintf(`{"socio":"s1","pista":%q,"inicio":"2026-03-10T18:00:00Z"}`, pista)
	resp, err := http.Post(url+"/reservas", "application/json", strings.NewReader(cuerpo))
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
}

// version la fijaría el proceso de compilación (go build -ldflags "-X main.version=...").
var version = "1.4.0"

func main() {
	fmt.Println("== Logs de tres peticiones ==")
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("version", version)
	srv := nuevoServidor(log, metricas.NuevoRegistro())
	for _, pista := range []string{"pista-1", "pista-2", "pista-9"} {
		reservar(srv.URL, pista)
	}
	srv.Close()

	fmt.Println("\n== Métricas de 200 peticiones ==")
	m := metricas.NuevoRegistro()
	srv = nuevoServidor(slog.New(slog.NewJSONHandler(io.Discard, nil)), m)
	for i := range 200 {
		pista := "pista-1"
		if i%10 == 0 {
			pista = "pista-2"
		}
		reservar(srv.URL, pista)
	}
	srv.Close()
	fmt.Print(m.Resumen())
	for _, p := range []float64{50, 95, 99} {
		fmt.Printf("p%v: %s\n", p, m.Percentil("POST /reservas", p).Round(time.Millisecond))
	}

	fmt.Println("\n== Alertas ==")
	if a, avisa := alertas.RevisarDuracion("recordatorios", 95*time.Minute, 40*time.Minute, 1.5); avisa {
		fmt.Println(a.Mensaje)
	}
	g := alertas.NuevoAgrupador(10 * time.Minute)
	t0 := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	limite := alertas.Alerta{Clave: "limite-api-pagos", Mensaje: "el proveedor de pagos devuelve 429"}
	for i := range 8 {
		ahora := t0.Add(time.Duration(i) * 2 * time.Minute)
		if enviar, calladas := g.Decidir(limite, ahora); enviar {
			fmt.Printf("%s enviar: %s (calladas desde el último aviso: %d)\n", ahora.Format("15:04"), limite.Mensaje, calladas)
		}
	}
}
