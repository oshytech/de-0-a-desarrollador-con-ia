import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Reservar } from "./Reservar";
import { crearApiSimulada } from "../../api/apiSimulada";

const ahora = () => new Date("2026-11-23T09:30:00+01:00");

async function elegirPistaYDia(dia: string) {
  const usuario = userEvent.setup();
  await usuario.selectOptions(await screen.findByLabelText("Pista"), "p-1");
  await usuario.selectOptions(screen.getByLabelText("Día"), dia);
  return usuario;
}

describe("Reservar", () => {
  it("reserva una hora libre y lo confirma de forma accesible", async () => {
    render(<Reservar api={crearApiSimulada(ahora)} socioId="s-1" ahora={ahora} />);
    const usuario = await elegirPistaYDia("2026-11-24");

    await usuario.click(await screen.findByRole("button", { name: "Reservar a las 10:00" }));

    expect(await screen.findByRole("status")).toHaveTextContent(
      "Reserva confirmada: Pista 1, martes, 24 de noviembre a las 10:00.",
    );
    // la hora reservada desaparece de la lista
    expect(screen.queryByRole("button", { name: "Reservar a las 10:00" })).not.toBeInTheDocument();
  });

  it("enseña el mensaje de la API cuando rechaza la reserva", async () => {
    const api = crearApiSimulada(ahora);
    // Otra persona ya tiene la pista 1 a las 10:00, pero esta pantalla aún no lo sabe.
    const horasViejas = await api.horasLibres("p-1", "2026-11-24");
    await api.reservar("s-3", "p-1", "2026-11-24T10:00");
    const apiConDatosViejos = { ...api, horasLibres: async () => horasViejas };

    render(<Reservar api={apiConDatosViejos} socioId="s-1" ahora={ahora} />);
    const usuario = await elegirPistaYDia("2026-11-24");
    await usuario.click(await screen.findByRole("button", { name: "Reservar a las 10:00" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Esa hora acaba de reservarla otra persona.");
  });

  it("avisa si el día elegido no tiene horas libres", async () => {
    const tarde = () => new Date("2026-11-23T21:30:00+01:00");
    render(<Reservar api={crearApiSimulada(tarde)} socioId="s-1" ahora={tarde} />);
    await elegirPistaYDia("2026-11-23");

    expect(await screen.findByText("No quedan horas libres ese día en esta pista.")).toBeInTheDocument();
  });
});
