// Package arquitectura contiene reglas de arquitectura escritas como tests.
package arquitectura

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// El dominio solo puede importar la librería estándar: ni adaptadores,
// ni aplicación, ni librerías de terceros.
func TestDominioSinDependencias(t *testing.T) {
	ficheros, err := filepath.Glob("../dominio/*.go")
	if err != nil || len(ficheros) == 0 {
		t.Fatalf("no encuentro el paquete dominio: %v", err)
	}
	fset := token.NewFileSet()
	for _, f := range ficheros {
		arbol, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range arbol.Imports {
			ruta, _ := strconv.Unquote(imp.Path.Value)
			primero := strings.Split(ruta, "/")[0]
			deTerceros := strings.Contains(primero, ".")
			delProyecto := primero == "reservas"
			if deTerceros || delProyecto {
				t.Errorf("%s importa %q: el dominio solo puede usar la librería estándar",
					filepath.Base(f), ruta)
			}
		}
	}
}
