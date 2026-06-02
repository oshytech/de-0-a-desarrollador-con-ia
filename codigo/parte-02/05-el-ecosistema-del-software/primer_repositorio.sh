#!/usr/bin/env bash
# Recorre el ciclo básico de Git del capítulo 5 en una carpeta temporal.
# Uso: bash primer_repositorio.sh
set -e
carpeta=$(mktemp -d)
cd "$carpeta"
echo "Trabajando en $carpeta"

git init -b main
git config user.name "Tu Nombre"
git config user.email "tu@correo.example"
git config core.autocrlf false

echo "# Mis apuntes" > README.md
echo "- Un algoritmo es una secuencia finita de pasos precisos." > conceptos.md
git status --short

git add README.md conceptos.md
git commit -m "Empieza los apuntes de conceptos"

echo "- El pseudocódigo se escribe para personas, no para máquinas." >> conceptos.md
git status --short
git diff

git add conceptos.md
git commit -m "Añade la definición de pseudocódigo"

git log --oneline
