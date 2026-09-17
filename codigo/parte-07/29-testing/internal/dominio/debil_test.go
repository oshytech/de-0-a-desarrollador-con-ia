package dominio

import "testing"

// Tests "débiles": pasan por todas las líneas, pero solo con valores cómodos,
// lejos de los límites. Sirven para comparar con la tabla en las pruebas de mutación.

func TestDebilHorario(t *testing.T) {
	if ComprobarHorario(en(1, 10, 0), ahora) != nil {
		t.Error("las 10:00 de mañana deberían valer")
	}
	if ComprobarHorario(en(1, 23, 0), ahora) == nil {
		t.Error("las 23:00 no deberían valer")
	}
	if ComprobarHorario(en(1, 10, 15), ahora) == nil {
		t.Error("una reserva que no empieza en punto no debería valer")
	}
	if ComprobarHorario(en(30, 10, 0), ahora) == nil {
		t.Error("a 30 días vista no debería valer")
	}
}

func TestDebilSolape(t *testing.T) {
	a := Reserva{Inicio: en(1, 10, 0), Fin: en(1, 11, 0)}
	if !SeSolapan(a, a) {
		t.Error("una reserva se solapa consigo misma")
	}
}

func TestDebilLimite(t *testing.T) {
	if ComprobarLimite(0) != nil || ComprobarLimite(5) == nil {
		t.Error("límite mal aplicado")
	}
}

func TestDebilCancelacion(t *testing.T) {
	if ComprobarCancelacion(Reserva{Inicio: en(0, 18, 0)}, ahora) != nil {
		t.Error("con horas de margen se debería poder cancelar")
	}
	if ComprobarCancelacion(Reserva{Inicio: en(0, 10, 0)}, ahora) == nil {
		t.Error("a media hora no se debería poder cancelar")
	}
}
