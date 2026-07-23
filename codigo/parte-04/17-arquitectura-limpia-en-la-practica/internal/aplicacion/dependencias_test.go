package aplicacion

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// La aplicación puede usar el dominio y la librería estándar, pero no conoce
// ningún adaptador ni ninguna tecnología concreta.
var prohibidasEnAplicacion = []string{"net/http", "database/sql", "encoding/json"}

func TestLaAplicacionSoloConoceElDominio(t *testing.T) {
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
			if strings.HasPrefix(ruta, "reservas/") && ruta != "reservas/internal/dominio" {
				t.Errorf("%s importa %q: la aplicación solo puede depender del dominio", fichero, ruta)
			}
			if strings.Contains(primerTramo, ".") || slices.Contains(prohibidasEnAplicacion, ruta) {
				t.Errorf("%s importa %q, que es un detalle técnico", fichero, ruta)
			}
		}
	}
}
