// Mutar es un probador de mutaciones mínimo para enseñar la idea.
// Copia el módulo a una carpeta temporal, aplica un cambio pequeño (un mutante)
// en internal/dominio/reglas.go y ejecuta los tests elegidos con -run.
// Si los tests fallan, el mutante muere. Si pasan, sobrevive.
//
// Uso: go run ./cmd/mutar -run Debil
//
//	go run ./cmd/mutar -run Tabla
//	go run ./cmd/mutar -run "Tabla|Frontera"
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type mutante struct {
	descripcion string
	original    string
	mutado      string
}

var mutantes = []mutante{
	{"apertura: < pasa a <=", "hora < HoraApertura", "hora <= HoraApertura"},
	{"cierre: > pasa a >=", "hora > UltimaHoraInicio", "hora >= UltimaHoraInicio"},
	{"plazo: > pasa a >=", "inicio.Sub(ahora) > AntelacionMaxima", "inicio.Sub(ahora) >= AntelacionMaxima"},
	{"solape: && pasa a ||", "a.Inicio.Before(b.Fin) && b.Inicio.Before(a.Fin)", "a.Inicio.Before(b.Fin) || b.Inicio.Before(a.Fin)"},
	{"límite: >= pasa a >", "confirmadasFuturas >= MaxReservasFuturas", "confirmadasFuturas > MaxReservasFuturas"},
	{"cancelación: < pasa a <=", "r.Inicio.Sub(ahora) < AntelacionMinimaCancelacion", "r.Inicio.Sub(ahora) <= AntelacionMinimaCancelacion"},
}

const objetivo = "internal/dominio/reglas.go"

func main() {
	patron := flag.String("run", "Tabla", "tests que se ejecutan contra cada mutante")
	flag.Parse()

	fuente, err := os.ReadFile(objetivo)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ejecútalo desde la raíz del módulo:", err)
		os.Exit(2)
	}

	muertos := 0
	for _, m := range mutantes {
		if !strings.Contains(string(fuente), m.original) {
			fmt.Fprintf(os.Stderr, "no encuentro %q en %s\n", m.original, objetivo)
			os.Exit(2)
		}
		dir := copiarModulo()
		mutada := strings.Replace(string(fuente), m.original, m.mutado, 1)
		escribir(filepath.Join(dir, objetivo), mutada)

		cmd := exec.Command("go", "test", "-count=1", "-run", *patron, "./internal/dominio/")
		cmd.Dir = dir
		resultado := "SOBREVIVE"
		if cmd.Run() != nil {
			resultado = "muere"
			muertos++
		}
		fmt.Printf("%-28s %s\n", m.descripcion, resultado)
		os.RemoveAll(dir)
	}
	fmt.Printf("\npuntuación de mutación con -run %s: %d de %d (%.0f %%)\n",
		*patron, muertos, len(mutantes), 100*float64(muertos)/float64(len(mutantes)))
}

func copiarModulo() string {
	dir, err := os.MkdirTemp("", "mutante-")
	if err != nil {
		panic(err)
	}
	for _, f := range []string{"go.mod", "internal/dominio/dominio.go", objetivo,
		"internal/dominio/reglas_test.go", "internal/dominio/debil_test.go"} {
		datos, err := os.ReadFile(f)
		if err != nil {
			panic(err)
		}
		escribir(filepath.Join(dir, f), string(datos))
	}
	return dir
}

func escribir(ruta, contenido string) {
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(ruta, []byte(contenido), 0o644); err != nil {
		panic(err)
	}
}
