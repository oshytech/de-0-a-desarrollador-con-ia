# Capítulo 36. Dar ojos a los agentes

Página de reservas del club (HTML y JavaScript sin frameworks) y tests de navegador con Playwright.

```bash
npm install
npx playwright install chromium   # solo si no tienes el navegador
npx playwright test               # arranca el servidor, ejecuta los tests y lo para
bash verificar.sh                 # la pirámide de verificación en un comando
npm start                         # para verla tú en http://127.0.0.1:4173
```

- `tests/reservas.spec.ts`: reservar, pista ocupada, formulario vacío, recorrido con teclado y comparación de captura.
- `tests/formulario-antes.spec.ts`: comprueba los defectos del formulario de ejemplo del capítulo 35.
- La captura de referencia está en `tests/reservas.spec.ts-snapshots/` y depende del sistema operativo. En otro sistema, genera la tuya con `npx playwright test --update-snapshots` y revísala antes de guardarla.
