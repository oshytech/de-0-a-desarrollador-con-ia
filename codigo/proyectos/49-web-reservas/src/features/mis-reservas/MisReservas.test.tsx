import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MisReservas } from "./MisReservas";
import { crearApiSimulada } from "../../api/apiSimulada";

const ahora = () => new Date("2026-11-23T09:30:00+01:00");

describe("MisReservas", () => {
  it("lista las reservas pendientes y permite cancelar con margen", async () => {
    const api = crearApiSimulada(ahora);
    await api.reservar("s-1", "p-1", "2026-11-24T10:00");

    render(<MisReservas api={api} socioId="s-1" ahora={ahora} />);
    expect(await screen.findByText("Pista 1, martes, 24 de noviembre a las 10:00")).toBeInTheDocument();
    await userEvent.click(
      await screen.findByRole("button", { name: "Cancelar la reserva del martes, 24 de noviembre a las 10:00" }),
    );

    expect(await screen.findByText("No tienes reservas pendientes.")).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("Reserva cancelada.");
  });

  it("no ofrece cancelar si faltan menos de 2 horas", async () => {
    const api = crearApiSimulada(ahora);
    await api.reservar("s-1", "p-1", "2026-11-23T11:00");

    render(<MisReservas api={api} socioId="s-1" ahora={ahora} />);

    expect(await screen.findByText("Ya no se puede cancelar (faltan menos de 2 horas).")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Cancelar/ })).not.toBeInTheDocument();
  });
});
