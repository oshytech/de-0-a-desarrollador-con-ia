import { describe, expect, it } from "vitest";
import { conDesfase, crearApiSimulada } from "./apiSimulada";

const ahora = () => new Date("2026-11-23T09:30:00+01:00");

describe("API simulada", () => {
  it("devuelve las horas con el desfase horario local, como la API", () => {
    expect(conDesfase("2026-11-24T08:00")).toBe("2026-11-24T08:00:00+01:00");
    expect(conDesfase("2026-07-01T08:00")).toBe("2026-07-01T08:00:00+02:00"); // horario de verano
  });

  it("no ofrece horas que ya han pasado", async () => {
    const horas = await crearApiSimulada(ahora).horasLibres("p-1", "2026-11-23");
    expect(horas[0].etiqueta).toBe("10:00");
    expect(horas.at(-1)?.etiqueta).toBe("21:00");
  });

  it("aplica las reglas del club", async () => {
    const api = crearApiSimulada(ahora);
    await api.reservar("s-1", "p-1", "2026-11-24T10:00");

    await expect(api.reservar("s-3", "p-1", "2026-11-24T10:00")).rejects.toMatchObject({ codigo: "pista_ocupada" });
    await expect(api.reservar("s-3", "p-1", "2026-11-24T10:30")).rejects.toMatchObject({ codigo: "horario_no_valido" });
    await expect(api.reservar("s-3", "p-1", "2026-12-08T10:00")).rejects.toMatchObject({ codigo: "fuera_de_plazo" });
    await expect(api.reservar("s-2", "p-2", "2026-11-24T10:00")).rejects.toMatchObject({ codigo: "socio_inactivo" });

    await api.reservar("s-1", "p-2", "2026-11-24T11:00");
    await expect(api.reservar("s-1", "p-2", "2026-11-24T12:00")).rejects.toMatchObject({ codigo: "limite_de_reservas" });
  });

  it("la reserva de otro socio se responde como si no existiera", async () => {
    const api = crearApiSimulada(ahora);
    const reserva = await api.reservar("s-1", "p-1", "2026-11-24T10:00");
    await expect(api.cancelar("s-3", reserva.id)).rejects.toMatchObject({ estado: 404 });
    expect((await api.cancelar("s-1", reserva.id)).estado).toBe("cancelada");
    await expect(api.cancelar("s-1", reserva.id)).rejects.toMatchObject({ codigo: "reserva_ya_cancelada" });
  });
});
