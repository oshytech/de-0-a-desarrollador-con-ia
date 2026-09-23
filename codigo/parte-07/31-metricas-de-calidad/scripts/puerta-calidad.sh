#!/bin/sh
# Puerta de calidad del proyecto. Si todo va bien no dice nada.
# Si algo falla, enseña solo lo que ha fallado y sale con error.
UMBRAL_COBERTURA=80
UMBRAL_COMPLEJIDAD=10

sin_formato=$(gofmt -l .)
if [ -n "$sin_formato" ]; then
	echo "Ficheros sin formatear (ejecuta gofmt -w .):"
	echo "$sin_formato"
	exit 1
fi

salida=$(go vet ./... 2>&1) || { echo "$salida"; exit 1; }

salida=$(go test -coverprofile=cobertura.out ./internal/... 2>&1) || { echo "$salida"; exit 1; }

total=$(go tool cover -func=cobertura.out | awk '/^total:/ {sub("%", "", $3); print $3}')
if awk -v t="$total" -v u="$UMBRAL_COBERTURA" 'BEGIN { exit !(t < u) }'; then
	echo "Cobertura del $total %, por debajo del $UMBRAL_COBERTURA %:"
	go tool cover -func=cobertura.out | awk '$3 != "100.0%"'
	exit 1
fi

salida=$(go run ./cmd/complejidad -umbral "$UMBRAL_COMPLEJIDAD" ./internal/dominio 2>&1) || { echo "$salida"; exit 1; }
