import { crearApiHttp } from "./apiHttp";
import { crearApiSimulada } from "./apiSimulada";
import type { ApiReservas } from "./tipos";

/** Elige la API según las variables de entorno de Vite (ver .env.example). */
export function crearApi(): ApiReservas {
  if (import.meta.env.VITE_API_SIMULADA === "true") {
    return crearApiSimulada(() => new Date(), 400);
  }
  return crearApiHttp(import.meta.env.VITE_API_URL ?? "/api");
}
