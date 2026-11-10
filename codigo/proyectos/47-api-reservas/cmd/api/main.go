// Composition root: aquí se eligen las piezas concretas y se conectan.
// No hay reglas de negocio en este fichero.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // incluye las zonas horarias en el binario

	"github.com/jackc/pgx/v5/pgxpool"

	"reservas/datos"
	apihttp "reservas/internal/adaptadores/http"
	"reservas/internal/adaptadores/memoria"
	"reservas/internal/adaptadores/postgres"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
	"reservas/migraciones"
)

func main() {
	if err := ejecutar(); err != nil {
		log.Fatal(err)
	}
}

func ejecutar() error {
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	zona, err := time.LoadLocation(valorPorDefecto(os.Getenv("ZONA_CLUB"), "Europe/Madrid"))
	if err != nil {
		return err
	}

	// Con RESERVAS_BD se usa PostgreSQL. Sin ella, los repositorios en memoria
	// con los datos de ejemplo: sirve para una demo, y al parar se pierde todo.
	var (
		socios   aplicacion.RepositorioSocios
		pistas   aplicacion.RepositorioPistas
		reservas aplicacion.RepositorioReservas
	)
	if url := os.Getenv("RESERVAS_BD"); url != "" {
		pool, err := pgxpool.New(ctx, url)
		if err != nil {
			return err
		}
		defer pool.Close()
		// Las migraciones se aplican antes de atender peticiones: si fallan,
		// esta versión no arranca (capítulo 37).
		aplicadas, err := postgres.Migrar(ctx, pool, migraciones.Ficheros)
		if err != nil {
			return err
		}
		log.Printf("migraciones aplicadas ahora: %v", aplicadas)
		if os.Getenv("DATOS_DE_EJEMPLO") == "si" {
			if _, err := pool.Exec(ctx, datos.Ejemplo); err != nil {
				return err
			}
			log.Print("datos de ejemplo cargados")
		}
		socios = postgres.NuevoRepositorioSocios(pool)
		pistas = postgres.NuevoRepositorioPistas(pool)
		reservas = postgres.NuevoRepositorioReservas(pool)
		log.Print("persistencia: PostgreSQL")
	} else {
		socios, pistas, reservas = memoria.DatosDeEjemplo()
		log.Print("persistencia: memoria (RESERVAS_BD vacía)")
	}

	relojDelClub := reloj.Sistema{}
	casos := apihttp.CasosDeUso{
		Reservar: aplicacion.NuevoReservarPista(reservas, socios, pistas, relojDelClub),
		Cancelar: aplicacion.NuevoCancelarReserva(reservas, relojDelClub),
		Listar:   aplicacion.NuevoListarReservasDeSocio(socios, reservas, relojDelClub),

		Pistas:         aplicacion.NuevoListarPistas(pistas),
		Disponibilidad: aplicacion.NuevoConsultarDisponibilidad(pistas, reservas, relojDelClub),
	}
	manejador := apihttp.NuevoManejador(casos, zona)

	// En tu ordenador se escucha solo en 127.0.0.1. Dentro de un contenedor
	// hace falta DIRECCION=0.0.0.0:8080 para que llegue el tráfico de fuera.
	servidor := &http.Server{
		Addr:              valorPorDefecto(os.Getenv("DIRECCION"), "127.0.0.1:8080"),
		Handler:           manejador.Rutas(),
		ReadHeaderTimeout: 5 * time.Second,  // sin estos límites, un cliente lento
		ReadTimeout:       10 * time.Second, // puede retener conexiones abiertas
		WriteTimeout:      10 * time.Second, // indefinidamente (capítulo 42)
		IdleTimeout:       60 * time.Second,
	}

	errores := make(chan error, 1)
	go func() {
		log.Printf("API de reservas escuchando en %s", servidor.Addr)
		errores <- servidor.ListenAndServe()
	}()

	select {
	case err := <-errores:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		log.Print("parando: se terminan las peticiones en curso")
		apagado, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelar()
		return servidor.Shutdown(apagado)
	}
}

func valorPorDefecto(valor, porDefecto string) string {
	if valor == "" {
		return porDefecto
	}
	return valor
}
