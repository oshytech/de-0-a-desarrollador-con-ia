import type { Reserva } from "../api/tipos";

// Reglas del club que la web usa para guiar al socio. La API es quien decide:
// aquí solo sirven para no ofrecer opciones que ya sabemos que van a fallar.

export const DIAS_DE_ANTELACION = 14; // regla 6
export const HORA_PRIMERA = 8; // regla 1
export const HORA_ULTIMA = 21; // regla 1: la última empieza a las 21:00
const MARGEN_CANCELACION_MS = 2 * 60 * 60 * 1000; // regla 5

/** Fecha local en formato AAAA-MM-DD, sin pasar por UTC. */
export function fechaLocal(fecha: Date): string {
  const mes = String(fecha.getMonth() + 1).padStart(2, "0");
  const dia = String(fecha.getDate()).padStart(2, "0");
  return `${fecha.getFullYear()}-${mes}-${dia}`;
}

/** Hoy y los 14 días siguientes. */
export function diasReservables(ahora: Date): string[] {
  return Array.from({ length: DIAS_DE_ANTELACION + 1 }, (_, i) => {
    const dia = new Date(ahora.getFullYear(), ahora.getMonth(), ahora.getDate() + i);
    return fechaLocal(dia);
  });
}

export function sePuedeCancelar(reserva: Reserva, ahora: Date): boolean {
  return (
    reserva.estado === "confirmada" &&
    Date.parse(reserva.inicio) - ahora.getTime() >= MARGEN_CANCELACION_MS
  );
}

const formatoDia = new Intl.DateTimeFormat("es-ES", { weekday: "long", day: "numeric", month: "long" });

/** "2026-11-24" → "martes, 24 de noviembre" */
export function nombreDelDia(fecha: string): string {
  const [anio, mes, dia] = fecha.split("-").map(Number);
  return formatoDia.format(new Date(anio, mes - 1, dia));
}
