// Package datos incrusta los datos de ejemplo. No es una migración:
// las migraciones se aplican en todos los entornos y esto solo en local.
package datos

import _ "embed"

//go:embed ejemplo.sql
var Ejemplo string
