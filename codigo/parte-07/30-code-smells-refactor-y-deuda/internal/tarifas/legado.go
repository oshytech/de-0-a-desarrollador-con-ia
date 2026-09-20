package tarifas

import "time"

// PrecioLegado es la versión original, con sus smells. Se conserva mientras dure
// el refactor para poder compararla con la nueva y se borrará al terminar.
func PrecioLegado(d string, inicio time.Time, luz bool, bono bool) float64 {
	p := 0.0
	if d == "tenis" {
		p = 12
		if inicio.Hour() >= 18 {
			p = p + 4
		}
		if inicio.Weekday() == time.Saturday || inicio.Weekday() == time.Sunday {
			p = p + 2
		}
		if luz {
			p = p + 3
		}
	} else if d == "padel" {
		p = 16
		if inicio.Hour() >= 18 {
			p = p + 4
		}
		if inicio.Weekday() == time.Saturday || inicio.Weekday() == time.Sunday {
			p = p + 2
		}
		if luz {
			p = p + 3
		}
	} else {
		return -1
	}
	if bono {
		p = p * 0.8
	}
	return p
}
