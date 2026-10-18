// API de reservas lista para ir en un contenedor.
//
//	api             arranca el servidor HTTP
//	api -migrar     aplica las migraciones pendientes y termina
//	api -comprobar  pide /vivo al servidor local (para el HEALTHCHECK)
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/jackc/pgx/v5/pgxpool"

	apihttp "reservas/internal/adaptadores/http"
	"reservas/internal/adaptadores/postgres"
	"reservas/migraciones"
)

func main() {
	migrar := flag.Bool("migrar", false, "aplica las migraciones y termina")
	comprobar := flag.Bool("comprobar", false, "comprueba que el servidor local responde")
	flag.Parse()

	// Fuera de un contenedor escuchamos solo en la propia máquina.
	// La imagen cambia la dirección a 0.0.0.0:8080 (ver el Dockerfile).
	direccion := entorno("DIRECCION", "127.0.0.1:8080")
	if *comprobar {
		os.Exit(comprobarSalud(direccion))
	}

	// La configuración llega por variables de entorno, nunca escrita en la imagen.
	url := os.Getenv("RESERVAS_BD")
	if url == "" {
		slog.Error("falta la variable RESERVAS_BD")
		os.Exit(1)
	}

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		slog.Error("conectar", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if *migrar {
		aplicadas, err := postgres.Migrar(ctx, pool, migraciones.Ficheros)
		if err != nil {
			slog.Error("migrar", "error", err)
			os.Exit(1)
		}
		slog.Info("migraciones aplicadas", "ficheros", aplicadas)
		return
	}

	srv := &http.Server{
		Addr:              direccion,
		Handler:           apihttp.NuevoMux(postgres.NuevoRepositorio(pool), pool),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		slog.Info("escuchando", "direccion", direccion)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("servidor", "error", err)
			os.Exit(1)
		}
	}()

	// Docker y Kubernetes avisan con SIGTERM antes de matar el contenedor:
	// dejamos terminar las peticiones en curso.
	<-ctx.Done()
	apagar, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	srv.Shutdown(apagar)
	slog.Info("apagado limpio")
}

func entorno(clave, porDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return porDefecto
}

func comprobarSalud(direccion string) int {
	_, puerto, err := net.SplitHostPort(direccion)
	if err != nil {
		return 1
	}
	cliente := http.Client{Timeout: 2 * time.Second}
	resp, err := cliente.Get("http://127.0.0.1:" + puerto + "/vivo")
	if err != nil || resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
