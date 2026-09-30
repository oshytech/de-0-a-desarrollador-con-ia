package dominio_test

import (
	"errors"
	"fmt"
	"time"

	"reservas/internal/dominio"
)

// Un martes cualquiera a las 10:00, que hace de "ahora" en todos los ejemplos.
var ahora = time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)

func ExampleNuevaReserva() {
	inicio := time.Date(2026, 3, 10, 18, 0, 0, 0, time.UTC)
	r, err := dominio.NuevaReserva("r1", "s1", "pista-1", inicio, ahora)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r.Inicio.Format("15:04"), "-", r.Fin.Format("15:04"), r.Estado)
	// Output: 18:00 - 19:00 confirmada
}

func ExampleNuevaReserva_horarioNoValido() {
	inicio := time.Date(2026, 3, 10, 18, 30, 0, 0, time.UTC)
	_, err := dominio.NuevaReserva("r1", "s1", "pista-1", inicio, ahora)
	fmt.Println(errors.Is(err, dominio.ErrHorarioNoValido))
	// Output: true
}

func ExampleReserva_SeSolapaCon() {
	a, _ := dominio.NuevaReserva("a", "s1", "pista-1", time.Date(2026, 3, 10, 18, 0, 0, 0, time.UTC), ahora)
	b, _ := dominio.NuevaReserva("b", "s2", "pista-1", time.Date(2026, 3, 10, 19, 0, 0, 0, time.UTC), ahora)
	fmt.Println(a.SeSolapaCon(b))
	// Output: false
}

func ExampleReserva_Cancelar() {
	r, _ := dominio.NuevaReserva("r1", "s1", "pista-1", time.Date(2026, 3, 10, 11, 0, 0, 0, time.UTC), ahora)
	err := r.Cancelar(ahora) // falta una hora para el inicio
	fmt.Println(err)
	// Output: solo se puede cancelar hasta 2 horas antes del inicio
}
