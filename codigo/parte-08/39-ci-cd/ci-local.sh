#!/usr/bin/env bash
# Ejecuta en tu máquina los mismos pasos que el job "pruebas" y la
# construcción de la imagen del workflow ci.yml.
# Uso: ./ci-local.sh [carpeta del código]   (por defecto, ../38-contenedores)
set -euo pipefail

codigo="${1:-../38-contenedores}"
cd "$codigo"

contenedor="reservas-ci-local"
docker run -d --rm --name "$contenedor" \
  -e POSTGRES_USER=club -e POSTGRES_PASSWORD=solo-para-ci -e POSTGRES_DB=club \
  -p 127.0.0.1:55433:5432 postgres:17-alpine > /dev/null
trap 'docker stop "$contenedor" > /dev/null' EXIT

until docker exec "$contenedor" pg_isready -U club > /dev/null 2>&1; do sleep 1; done

echo "== Formato";           test -z "$(gofmt -l .)"
echo "== Análisis estático"; go vet ./...
echo "== Tests";             RESERVAS_BD="postgres://club:solo-para-ci@localhost:55433/club?sslmode=disable" go test ./...
echo "== Imagen";            docker build -q -t reservas-api:local . > /dev/null
echo "Todo en verde"
