import { describe, expect, it, vi } from "vitest";
import { crearApiHttp } from "./apiHttp";
import { ErrorApi } from "./tipos";

function respuesta(estado: number, cuerpo?: unknown): Response {
  return new Response(cuerpo === undefined ? null : JSON.stringify(cuerpo), {
    status: estado,
    headers: { "Content-Type": "application/json" },
  });
}

const reservaJson = {
  id: "r-7",
  socio_id: "s-1",
  pista_id: "p-2",
  inicio: "2026-11-24T10:00:00+01:00",
  fin: "2026-11-24T11:00:00+01:00",
  estado: "confirmada",
};

describe("crearApiHttp", () => {
  it("convierte la disponibilidad en horas listas para reservar", async () => {
    const fetchFalso = vi.fn().mockResolvedValue(respuesta(200, { fecha: "2026-11-24", horas_libres: ["08:00", "09:00"] }));

    const horas = await crearApiHttp("/api", fetchFalso).horasLibres("p-1", "2026-11-24");

    expect(fetchFalso).toHaveBeenCalledWith("/api/pistas/p-1/disponibilidad?fecha=2026-11-24", {});
    expect(horas).toEqual([
      { inicio: "2026-11-24T08:00", etiqueta: "08:00" },
      { inicio: "2026-11-24T09:00", etiqueta: "09:00" },
    ]);
  });

  it("reserva como el socio de la cabecera y traduce la respuesta", async () => {
    const fetchFalso = vi.fn().mockResolvedValue(respuesta(201, reservaJson));

    const reserva = await crearApiHttp("/api", fetchFalso).reservar("s-1", "p-2", "2026-11-24T10:00");

    const [url, opciones] = fetchFalso.mock.calls[0];
    expect(url).toBe("/api/reservas");
    expect(opciones.method).toBe("POST");
    expect(opciones.headers["X-Socio-Id"]).toBe("s-1");
    expect(JSON.parse(opciones.body)).toEqual({ pista_id: "p-2", inicio: "2026-11-24T10:00" });
    expect(reserva).toMatchObject({ id: "r-7", socioId: "s-1", pistaId: "p-2" });
  });

  it("un 409 se convierte en ErrorApi con el mensaje de la API", async () => {
    const fetchFalso = vi
      .fn()
      .mockResolvedValue(respuesta(409, { error: "pista_ocupada", mensaje: "La pista ya está reservada a esa hora." }));

    const promesa = crearApiHttp("/api", fetchFalso).reservar("s-1", "p-1", "2026-11-24T10:00");

    await expect(promesa).rejects.toBeInstanceOf(ErrorApi);
    await expect(promesa).rejects.toMatchObject({ estado: 409, codigo: "pista_ocupada" });
    await expect(promesa).rejects.toThrow("La pista ya está reservada a esa hora.");
  });

  it("si el error no trae cuerpo JSON, usa un mensaje por defecto", async () => {
    const fetchFalso = vi.fn().mockResolvedValue(new Response("Unprocessable", { status: 422 }));
    await expect(crearApiHttp("/api", fetchFalso).pistas()).rejects.toThrow("La reserva no cumple las normas del club.");
  });

  it("cancelar usa la ruta de cancelación con la cabecera del socio", async () => {
    const fetchFalso = vi.fn().mockResolvedValue(respuesta(200, { ...reservaJson, estado: "cancelada" }));

    const reserva = await crearApiHttp("/api", fetchFalso).cancelar("s-1", "r-7");

    expect(fetchFalso).toHaveBeenCalledWith("/api/reservas/r-7/cancelacion", {
      method: "POST",
      headers: { "X-Socio-Id": "s-1" },
    });
    expect(reserva.estado).toBe("cancelada");
  });

  it("un fallo de red se convierte en un error legible", async () => {
    const fetchFalso = vi.fn().mockRejectedValue(new TypeError("Failed to fetch"));
    await expect(crearApiHttp("/api", fetchFalso).pistas()).rejects.toMatchObject({ codigo: "sin_conexion" });
  });
});
