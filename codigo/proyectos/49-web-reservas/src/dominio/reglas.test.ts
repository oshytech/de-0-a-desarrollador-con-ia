import { describe, expect, it } from "vitest";
import { diasReservables, nombreDelDia, sePuedeCancelar } from "./reglas";
import type { Reserva } from "../api/tipos";

const ahora = new Date("2026-11-23T09:30:00+01:00"); // lunes

function reserva(inicio: string, estado: Reserva["estado"] = "confirmada"): Reserva {
  return { id: "r-1", socioId: "s-1", pistaId: "p-1", inicio, fin: inicio, estado };
}

describe("diasReservables", () => {
  it("ofrece hoy y los 14 días siguientes", () => {
    const dias = diasReservables(ahora);
    expect(dias).toHaveLength(15);
    expect(dias[0]).toBe("2026-11-23");
    expect(dias.at(-1)).toBe("2026-12-07");
  });

  it("usa la fecha local aunque en UTC ya sea otro día", () => {
    // 00:30 en Madrid son las 23:30 del día anterior en UTC
    expect(diasReservables(new Date("2026-11-24T00:30:00+01:00"))[0]).toBe("2026-11-24");
  });
});

describe("sePuedeCancelar", () => {
  it("permite cancelar con 2 horas o más de margen", () => {
    expect(sePuedeCancelar(reserva("2026-11-23T11:30:00+01:00"), ahora)).toBe(true);
  });

  it("no permite cancelar con menos de 2 horas", () => {
    expect(sePuedeCancelar(reserva("2026-11-23T11:00:00+01:00"), ahora)).toBe(false);
  });

  it("no permite cancelar una reserva ya cancelada", () => {
    expect(sePuedeCancelar(reserva("2026-11-25T10:00:00+01:00", "cancelada"), ahora)).toBe(false);
  });
});

describe("nombreDelDia", () => {
  it("escribe el día en español", () => {
    expect(nombreDelDia("2026-11-24")).toBe("martes, 24 de noviembre");
  });
});
