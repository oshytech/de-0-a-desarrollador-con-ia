#!/usr/bin/env bash
# Hábitos de Git del capítulo 9 en una carpeta temporal: .gitignore, ramas y fusión.
# Uso: bash git_habitos.sh (desde esta carpeta)
set -e
origen=$(pwd)
carpeta=$(mktemp -d)
cd "$carpeta"
git init -q -b main
git config user.name "Tu Nombre"
git config user.email "tu@correo.example"
git config core.autocrlf false

# 1. Lo primero, el .gitignore
mkdir -p .venv __pycache__
touch .venv/python __pycache__/entrenamientos.cpython-313.pyc .env app.log
cp "$origen/entrenamientos.py" .
echo "--- git status sin .gitignore ---"
git status --short
cp "$origen/ejemplo.gitignore" .gitignore
echo "--- git status con .gitignore ---"
git status --short
git add .gitignore entrenamientos.py
git commit -q -m "Añade el resumen semanal de entrenamientos"

# 2. Una rama para un cambio
git switch -q -c ritmo-medio-semanal
sed -i 's/Ritmo medio:/Ritmo medio semanal:/' entrenamientos.py
git commit -q -am "Aclara que el ritmo medio es semanal"
git switch -q main
git merge -q ritmo-medio-semanal
git branch -d ritmo-medio-semanal
echo "--- historial ---"
git log --oneline

# 3. Deshacer un cambio que aún no has guardado
echo "basura" >> entrenamientos.py
git status --short
git restore entrenamientos.py
git status --short
echo "--- limpio ---"
