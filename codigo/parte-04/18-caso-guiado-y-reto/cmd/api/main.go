// Composition root: aquí se eligen las piezas concretas y se conectan.
// No hay reglas de negocio en este fichero.
package main

import (
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata" // incluye las zonas horarias en el binario

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

	casos := apihttp.CasosDeUso{
		Reservar: aplicacion.NuevoReservarPista(reservas, socios, pistas, relojDelClub),
		Cancelar: aplicacion.NuevoCancelarReserva(reservas, relojDelClub),
		Listar:   aplicacion.NuevoListarReservasDeSocio(socios, reservas, relojDelClub),
	}
	manejador := apihttp.NuevoManejador(casos, zona)

	// En local se escucha solo en 127.0.0.1. Dentro de un contenedor habría
	// que usar DIRECCION=0.0.0.0:8080.
	direccion := valorPorDefecto(os.Getenv("DIRECCION"), "127.0.0.1:8080")
	log.Printf("API de reservas escuchando en %s", direccion)
	log.Fatal(http.ListenAndServe(direccion, manejador.Rutas()))
}

func valorPorDefecto(valor, porDefecto string) string {
	if valor == "" {
		return porDefecto
	}
	return valor
}
