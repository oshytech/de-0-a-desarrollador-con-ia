// Command complejidad calcula la complejidad ciclomática de cada función de un
// directorio Go: 1 más cada punto donde el flujo puede tomar otro camino.
//
//	go run ./cmd/complejidad -umbral 10 ./internal/dominio
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type medida struct {
	funcion     string
	posicion    string
	complejidad int
}

func complejidad(cuerpo ast.Node) int {
	total := 1
	ast.Inspect(cuerpo, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			total++
		case *ast.CaseClause:
			if x.List != nil { // el default no suma
				total++
			}
		case *ast.CommClause:
			if x.Comm != nil {
				total++
			}
		case *ast.BinaryExpr:
			if x.Op == token.LAND || x.Op == token.LOR {
				total++
			}
		}
		return true
	})
	return total
}

func main() {
	umbral := flag.Int("umbral", 10, "complejidad a partir de la cual se marca la función")
	flag.Parse()
	dir := flag.Arg(0)
	if dir == "" {
		dir = "."
	}

	fset := token.NewFileSet()
	ficheros, _ := filepath.Glob(filepath.Join(dir, "*.go"))
	var medidas []medida
	for _, f := range ficheros {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		arbol, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		for _, d := range arbol.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Body != nil {
				p := fset.Position(fn.Pos())
				medidas = append(medidas, medida{
					funcion:     fn.Name.Name,
					posicion:    fmt.Sprintf("%s:%d", filepath.Base(p.Filename), p.Line),
					complejidad: complejidad(fn.Body),
				})
			}
		}
	}

	sort.Slice(medidas, func(i, j int) bool { return medidas[i].complejidad > medidas[j].complejidad })
	superan := 0
	for _, m := range medidas {
		marca := ""
		if m.complejidad > *umbral {
			marca = "  <- revisar"
			superan++
		}
		fmt.Printf("%3d  %-16s %s%s\n", m.complejidad, m.funcion, m.posicion, marca)
	}
	if superan > 0 {
		fmt.Printf("%d funciones superan el umbral de %d\n", superan, *umbral)
		os.Exit(1)
	}
}
