// Orientación a objetos al estilo de Go: structs, métodos e interfaces implícitas.
package main

import (
	"errors"
	"fmt"
	"time"
)

const margenCancelacion = 2 * time.Hour

type Reserva struct {
	ID     int
	Socio  string
	Pista  string
	Inicio time.Time
	estado string // en minúscula: no se ve fuera del paquete
}

func NuevaReserva(id int, socio, pista string, inicio time.Time) *Reserva {
	return &Reserva{ID: id, Socio: socio, Pista: pista, Inicio: inicio, estado: "confirmada"}
}

func (r *Reserva) Estado() string { return r.estado }

func (r *Reserva) Cancelar(ahora time.Time) error {
	if r.estado == "cancelada" {
		return errors.New("la reserva ya estaba cancelada")
	}
	if r.Inicio.Sub(ahora) < margenCancelacion {
		return errors.New("solo se puede cancelar hasta 2 horas antes")
	}
	r.estado = "cancelada"
	return nil
}

// Notificador es una interfaz: cualquier tipo con este método la cumple,
// sin escribir "implements" en ninguna parte.
type Notificador interface {
	Notificar(socio, mensaje string) error
}

type Correo struct{}

func (Correo) Notificar(socio, mensaje string) error {
	fmt.Printf("[correo a %s] %s\n", socio, mensaje)
	return nil
}

type SMS struct{ Remitente string }

func (s SMS) Notificar(socio, mensaje string) error {
	fmt.Printf("[SMS de %s a %s] %s\n", s.Remitente, socio, mensaje)
	return nil
}

func avisarCancelacion(n Notificador, r *Reserva) error {
	return n.Notificar(r.Socio, fmt.Sprintf("Cancelada tu reserva de %s a las %s", r.Pista, r.Inicio.Format("15:04")))
}

func main() {
	r := NuevaReserva(1, "Lucía", "Pádel 1", time.Date(2026, 3, 10, 19, 0, 0, 0, time.UTC))
	if err := r.Cancelar(time.Date(2026, 3, 10, 17, 30, 0, 0, time.UTC)); err != nil {
		fmt.Println("No:", err)
	}
	if err := r.Cancelar(time.Date(2026, 3, 10, 16, 0, 0, 0, time.UTC)); err == nil {
		fmt.Println(r.Estado())
	}
	for _, n := range []Notificador{Correo{}, SMS{Remitente: "Club"}} {
		if err := avisarCancelacion(n, r); err != nil {
			fmt.Println("No se pudo avisar:", err)
		}
	}
}
