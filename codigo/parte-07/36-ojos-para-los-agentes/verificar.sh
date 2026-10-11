#!/usr/bin/env bash
# La pirámide de verificación como un solo comando, de lo más barato a lo más caro.
# Si una etapa falla, el script se para ahí y devuelve un código distinto de cero.
set -euo pipefail

echo "== [1/3] Sintaxis =="
node --check servidor.mjs
node --check public/app.js

echo "== [2/3] Navegador: flujos, errores, teclado y aspecto =="
npx playwright test

echo "== [3/3] Revisión con IA contra las reglas del equipo =="
if [ -n "${REVISION_IA:-}" ]; then
  "$REVISION_IA"   # por ejemplo, el script de revisión del capítulo 32
else
  echo "Sin revisor configurado (REVISION_IA): se salta esta etapa."
fi

echo "== Todo en verde =="
