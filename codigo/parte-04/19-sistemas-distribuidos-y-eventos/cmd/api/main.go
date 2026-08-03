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
	"time"
	_ "time/tzdata" // incluye las zonas horarias en el binario

	"reservas/internal/adaptadores/correo"
	"reservas/internal/adaptadores/eventos"
	apihttp "reservas/internal/adaptadores/http"
	"reservas/internal/adaptadores/memoria"
	"reservas/internal/adaptadores/reloj"
	"reservas/internal/aplicacion"
)

func main() {
	zona, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.Fatal(err)
	}
	socios, pistas, reservas := memoria.DatosDeEjemplo()
	relojDelClub := reloj.Sistema{}

	// El bus y sus suscriptores. Añadir otra reacción es añadir otra línea aquí,
	// sin tocar ReservarPista ni CancelarReserva.
	bus := eventos.NuevoBus(3, 200*time.Millisecond)
	avisar := aplicacion.NuevoAvisarAlSocio(socios, correo.NuevaConsola(os.Stdout))
	bus.Suscribir("avisar-al-socio", avisar.Manejar)

	casos := apihttp.CasosDeUso{
		Reservar: aplicacion.NuevoReservarPista(reservas, socios, pistas, relojDelClub, bus),
		Cancelar: aplicacion.NuevoCancelarReserva(reservas, relojDelClub, bus),
		Listar:   aplicacion.NuevoListarReservasDeSocio(socios, reservas, relojDelClub),
	}

	// En local se escucha solo en 127.0.0.1. Dentro de un contenedor habría
	// que usar DIRECCION=0.0.0.0:8080.
	servidor := &http.Server{
		Addr:    valorPorDefecto(os.Getenv("DIRECCION"), "127.0.0.1:8080"),
		Handler: apihttp.NuevoManejador(casos, zona).Rutas(),
	}

	// Parada ordenada con Ctrl+C: primero deja de aceptar peticiones,
	// después vacía el bus para no perder los avisos pendientes.
	senal, parar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer parar()
	go func() {
		<-senal.Done()
		apagado, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelar()
		servidor.Shutdown(apagado)
	}()

	log.Printf("API de reservas escuchando en %s", servidor.Addr)
	if err := servidor.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	bus.Cerrar()
	log.Printf("parada completa, eventos fallidos: %d", len(bus.Fallidos()))
}

func valorPorDefecto(valor, porDefecto string) string {
	if valor == "" {
		return porDefecto
	}
	return valor
}
