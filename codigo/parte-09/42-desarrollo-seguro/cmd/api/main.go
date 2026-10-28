// Composición del ejemplo: carga la configuración (o se niega a arrancar)
// y expone POST /reservas.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	adaptadorhttp "desarrolloseguro/internal/adaptadores/http"
	"desarrolloseguro/internal/config"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Cargar(os.Getenv)
	if err != nil {
		log.Error("no arranco", "motivo", err.Error())
		os.Exit(1)
	}
	log.Info("configuración cargada", "config", cfg) // los secretos salen como [oculto]

	zona, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.Error("zona horaria", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("POST /reservas", adaptadorhttp.ManejadorReservar{Ahora: time.Now, ZonaClub: zona, Log: log})

	srv := &http.Server{
		Addr:              cfg.Direccion,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	log.Info("escuchando", "direccion", cfg.Direccion)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("servidor parado", "error", err)
		os.Exit(1)
	}
}
