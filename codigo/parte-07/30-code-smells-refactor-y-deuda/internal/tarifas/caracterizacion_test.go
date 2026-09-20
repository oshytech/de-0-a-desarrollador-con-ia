package tarifas

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

var actualizar = flag.Bool("actualizar", false, "regenera testdata/precios.golden desde PrecioLegado")

const rutaGolden = "testdata/precios.golden"

type caso struct {
	deporte   string
	inicio    time.Time
	luz, bono bool
}

func (c caso) clave() string {
	return fmt.Sprintf("%s %s luz=%t bono=%t", c.deporte, c.inicio.Format("Mon 15:04"), c.luz, c.bono)
}

// todosLosCasos recorre cada deporte (incluido uno que no existe), cada hora
// permitida de una semana entera y las cuatro combinaciones de extras.
func todosLosCasos() []caso {
	var casos []caso
	lunes := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	for _, d := range []string{"tenis", "padel", "squash"} {
		for dia := 0; dia < 7; dia++ {
			for hora := 8; hora <= 21; hora++ {
				inicio := lunes.AddDate(0, 0, dia).Add(time.Duration(hora) * time.Hour)
				for _, luz := range []bool{false, true} {
					for _, bono := range []bool{false, true} {
						casos = append(casos, caso{d, inicio, luz, bono})
					}
				}
			}
		}
	}
	return casos
}

// TestCaracterizacionLegado fija lo que hace hoy el código antiguo, sin juzgarlo.
func TestCaracterizacionLegado(t *testing.T) {
	var b strings.Builder
	for _, c := range todosLosCasos() {
		fmt.Fprintf(&b, "%s => %.2f\n", c.clave(), PrecioLegado(c.deporte, c.inicio, c.luz, c.bono))
	}
	if *actualizar {
		if err := os.WriteFile(rutaGolden, []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	esperado, err := os.ReadFile(rutaGolden)
	if err != nil {
		t.Fatalf("no hay fichero golden, ejecuta go test -run Caracterizacion -actualizar: %v", err)
	}
	if b.String() != string(esperado) {
		t.Fatal("PrecioLegado ya no se comporta como dice testdata/precios.golden")
	}
}

// Divergencias decididas a propósito: el resto de casos tiene que coincidir.
var divergencias = map[string]string{
	"squash": "el legado devuelve -1; la versión nueva devuelve ErrDeporteDesconocido",
}

// TestParidad compara la versión nueva con la antigua en todos los casos.
func TestParidad(t *testing.T) {
	distintos := 0
	for _, c := range todosLosCasos() {
		viejo := PrecioLegado(c.deporte, c.inicio, c.luz, c.bono)
		nuevo, err := Precio(Deporte(c.deporte), c.inicio, Extras{ConLuz: c.luz, ConBono: c.bono})

		if _, decidida := divergencias[c.deporte]; decidida {
			if viejo != -1 || !errors.Is(err, ErrDeporteDesconocido) {
				t.Errorf("%s: la divergencia registrada ya no se cumple", c.clave())
			}
			continue
		}
		if err != nil || Centimos(math.Round(viejo*100)) != nuevo {
			distintos++
			if distintos <= 5 {
				t.Errorf("%s: legado %.2f, nuevo %d céntimos (err %v)", c.clave(), viejo, nuevo, err)
			}
		}
	}
	if distintos > 0 {
		t.Errorf("%d casos distintos", distintos)
	}
}
