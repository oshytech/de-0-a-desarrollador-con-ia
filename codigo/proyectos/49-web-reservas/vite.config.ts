import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// En desarrollo, /api se reenvía a la API del capítulo 47 (puerto 8080).
// Así la web y la API comparten origen y no hace falta configurar CORS.
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        rewrite: (ruta) => ruta.replace(/^\/api/, ""),
      },
    },
  },
});
