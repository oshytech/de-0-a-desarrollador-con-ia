import { defineConfig, devices } from "@playwright/test";

const host = process.env.HOST ?? "127.0.0.1";
const puerto = process.env.PORT ?? "4173";
const direccion = `http://${host}:${puerto}`;

export default defineConfig({
  testDir: "tests",
  fullyParallel: true,
  reporter: "list",
  use: {
    baseURL: direccion,
    locale: "es-ES",
    // Evidencia solo cuando algo falla: captura y traza para mirar qué vio el navegador.
    screenshot: "only-on-failure",
    trace: "retain-on-failure",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  // Playwright arranca el servidor antes de los tests y lo para al terminar.
  webServer: {
    command: "node servidor.mjs",
    url: direccion,
    reuseExistingServer: false,
  },
});
