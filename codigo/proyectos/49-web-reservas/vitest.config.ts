import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

// Zona horaria fija: los tests de fechas dan lo mismo en cualquier ordenador.
process.env.TZ = "Europe/Madrid";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
  },
});
