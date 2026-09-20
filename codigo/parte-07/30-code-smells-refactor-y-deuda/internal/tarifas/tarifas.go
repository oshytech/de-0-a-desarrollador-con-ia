// Package tarifas calcula lo que paga un socio por una hora de pista.
// Los importes son un ejemplo para el libro, no tarifas reales.
package tarifas

import (
	"errors"
	"time"
)

// Centimos evita los errores de redondeo de float64 con el dinero.
type Centimos int

type Deporte string

const (
	Tenis Deporte = "tenis"
	Padel Deporte = "padel"
)

var ErrDeporteDesconocido = errors.New("deporte desconocido")

var tarifaBase = map[Deporte]Centimos{
	Tenis: 1200,
	Padel: 1600,
}

const (
	InicioHoraPunta         = 18
	RecargoHoraPunta        = Centimos(400)
	RecargoFinDeSemana      = Centimos(200)
	RecargoLuz              = Centimos(300)
	DescuentoBonoPorcentaje = 20
)

// Extras sustituye a los dos parámetros booleanos sueltos de la versión antigua.
type Extras struct {
	ConLuz  bool
	ConBono bool
}

func Precio(d Deporte, inicio time.Time, extras Extras) (Centimos, error) {
	base, ok := tarifaBase[d]
	if !ok {
		return 0, ErrDeporteDesconocido
	}
	precio := base + recargos(inicio, extras)
	if extras.ConBono {
		precio = precio * (100 - DescuentoBonoPorcentaje) / 100
	}
	return precio, nil
}

func recargos(inicio time.Time, extras Extras) Centimos {
	var total Centimos
	if inicio.Hour() >= InicioHoraPunta {
		total += RecargoHoraPunta
	}
	if esFinDeSemana(inicio) {
		total += RecargoFinDeSemana
	}
	if extras.ConLuz {
		total += RecargoLuz
	}
	return total
}

func esFinDeSemana(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}
