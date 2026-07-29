package dominio

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// El dominio solo puede usar la librería estándar, y ni siquiera toda:
// nada de red, base de datos, JSON ni sistema operativo.
var prohibidasEnElDominio = []string{"net/http", "database/sql", "encoding/json", "os"}

func TestElDominioNoDependeDeDetalles(t *testing.T) {
	ficheros, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	conjunto := token.NewFileSet()
	for _, fichero := range ficheros {
		if strings.HasSuffix(fichero, "_test.go") {
			continue
		}
		arbol, err := parser.ParseFile(conjunto, fichero, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range arbol.Imports {
			ruta, _ := strconv.Unquote(imp.Path.Value)
			primerTramo := strings.Split(ruta, "/")[0]
			if strings.HasPrefix(ruta, "reservas/") || strings.Contains(primerTramo, ".") {
				t.Errorf("%s importa %q: el dominio no depende de nada del proyecto ni de fuera", fichero, ruta)
			}
			if slices.Contains(prohibidasEnElDominio, ruta) {
				t.Errorf("%s importa %q, que es un detalle técnico", fichero, ruta)
			}
		}
	}
}
