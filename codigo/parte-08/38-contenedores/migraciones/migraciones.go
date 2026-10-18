// Package migraciones incrusta los ficheros SQL en el binario,
// para que el programa no dependa de encontrarlos en disco.
package migraciones

import "embed"

//go:embed *.sql
var Ficheros embed.FS
