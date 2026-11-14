import { aHoraLibre } from "./apiHttp";
import { ErrorApi, type ApiReservas, type Pista, type Reserva } from "./tipos";
import { DIAS_DE_ANTELACION, HORA_PRIMERA, HORA_ULTIMA, diasReservables, sePuedeCancelar } from "../dominio/reglas";

// Los mismos datos de ejemplo que la API del capítulo 47.
const PISTAS: Pista[] = [
  { id: "p-1", nombre: "Pista 1", deporte: "tenis" },
  { id: "p-2", nombre: "Pista 2", deporte: "padel" },
];
const SOCIOS_INACTIVOS = new Set(["s-2"]);

/** "2026-11-24T10:00" → "2026-11-24T10:00:00+01:00", como responde la API. */
export function conDesfase(local: string): string {
  const minutos = -new Date(local).getTimezoneOffset();
  const signo = minutos >= 0 ? "+" : "-";
  const hh = String(Math.floor(Math.abs(minutos) / 60)).padStart(2, "0");
  const mm = String(Math.abs(minutos) % 60).padStart(2, "0");
  return `${local}:00${signo}${hh}:${mm}`;
}

const dosCifras = (n: number) => String(n).padStart(2, "0");

/**
 * API de mentira en memoria con las reglas del club, para desarrollar la web sin
 * levantar el servidor. Puede simular espera para ver los estados de carga.
 */
export function crearApiSimulada(ahora: () => Date = () => new Date(), esperaMs = 0): ApiReservas {
  const reservas: Reserva[] = [];
  let siguienteId = 1;
  const esperar = () => new Promise((r) => setTimeout(r, esperaMs));
  const fallo = (estado: number, codigo: string, mensaje: string) => new ErrorApi(estado, codigo, mensaje);
  const ocupada = (pistaId: string, local: string) =>
    reservas.some((r) => r.pistaId === pistaId && r.estado === "confirmada" && r.inicio.startsWith(local));

  return {
    async pistas() {
      await esperar();
      return PISTAS;
    },

    async horasLibres(pistaId, fecha) {
      await esperar();
      const horas: string[] = [];
      for (let h = HORA_PRIMERA; h <= HORA_ULTIMA; h++) {
        const hora = `${dosCifras(h)}:00`;
        const local = `${fecha}T${hora}`;
        if (new Date(local) > ahora() && !ocupada(pistaId, local)) horas.push(hora);
      }
      return horas.map((hora) => aHoraLibre(fecha, hora));
    },

    async reservar(socioId, pistaId, inicio) {
      await esperar();
      const momento = new Date(inicio); // sin zona: JavaScript la interpreta como hora local
      if (SOCIOS_INACTIVOS.has(socioId)) {
        throw fallo(422, "socio_inactivo", "Tu cuenta no está activa. Pasa por recepción.");
      }
      if (momento.getMinutes() !== 0 || momento.getHours() < HORA_PRIMERA || momento.getHours() > HORA_ULTIMA) {
        throw fallo(422, "horario_no_valido", "Las reservas empiezan en punto, de 08:00 a 21:00.");
      }
      if (momento <= ahora() || !diasReservables(ahora()).includes(inicio.slice(0, 10))) {
        throw fallo(422, "fuera_de_plazo", `Solo se puede reservar desde ahora hasta dentro de ${DIAS_DE_ANTELACION} días.`);
      }
      if (ocupada(pistaId, inicio)) {
        throw fallo(409, "pista_ocupada", "Esa hora acaba de reservarla otra persona.");
      }
      const pendientes = reservas.filter(
        (r) => r.socioId === socioId && r.estado === "confirmada" && Date.parse(r.inicio) > ahora().getTime(),
      );
      if (pendientes.length >= 2) {
        throw fallo(422, "limite_de_reservas", "Ya tienes 2 reservas pendientes. Cancela una para reservar otra.");
      }
      const fin = `${inicio.slice(0, 11)}${dosCifras(momento.getHours() + 1)}:00`;
      const reserva: Reserva = {
        id: `r-${siguienteId++}`,
        socioId,
        pistaId,
        inicio: conDesfase(inicio),
        fin: conDesfase(fin),
        estado: "confirmada",
      };
      reservas.push(reserva);
      return { ...reserva };
    },

    async cancelar(socioId, reservaId) {
      await esperar();
      const reserva = reservas.find((r) => r.id === reservaId);
      // La reserva de otro socio se trata como si no existiera, igual que en la API real.
      if (!reserva || reserva.socioId !== socioId) throw fallo(404, "no_encontrado", "Esa reserva no existe.");
      if (reserva.estado === "cancelada") throw fallo(409, "reserva_ya_cancelada", "Esa reserva ya estaba cancelada.");
      if (!sePuedeCancelar(reserva, ahora())) {
        throw fallo(422, "cancelacion_fuera_de_plazo", "Solo se puede cancelar hasta 2 horas antes del inicio.");
      }
      reserva.estado = "cancelada";
      return { ...reserva };
    },

    async reservasDe(socioId) {
      await esperar();
      return reservas
        .filter((r) => r.socioId === socioId && r.estado === "confirmada" && Date.parse(r.inicio) > ahora().getTime())
        .sort((a, b) => Date.parse(a.inicio) - Date.parse(b.inicio))
        .map((r) => ({ ...r }));
    },
  };
}
