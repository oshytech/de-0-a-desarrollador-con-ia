// Tipos que usa la web. Son los del dominio del club, no una copia literal del JSON:
// la traducción entre los dos vive en apiHttp.ts.

export type Deporte = "tenis" | "padel";

export interface Pista {
  id: string; // "p-1"
  nombre: string;
  deporte: Deporte;
}

/** Una hora libre. `inicio` es el valor exacto que la API espera al reservar. */
export interface HoraLibre {
  inicio: string; // "2026-11-24T10:00", hora local del club
  etiqueta: string; // "10:00"
}

export interface Reserva {
  id: string;
  socioId: string;
  pistaId: string;
  inicio: string; // "2026-11-24T10:00:00+01:00"
  fin: string;
  estado: "confirmada" | "cancelada";
}

/** Error que la API devuelve con su código y un mensaje para el socio. */
export class ErrorApi extends Error {
  readonly estado: number;
  readonly codigo: string;

  constructor(estado: number, codigo: string, mensaje: string) {
    super(mensaje);
    this.name = "ErrorApi";
    this.estado = estado;
    this.codigo = codigo;
  }
}

/** Lo que la web necesita de la API. Hay dos implementaciones: HTTP y simulada. */
export interface ApiReservas {
  pistas(): Promise<Pista[]>;
  horasLibres(pistaId: string, fecha: string): Promise<HoraLibre[]>;
  reservar(socioId: string, pistaId: string, inicio: string): Promise<Reserva>;
  cancelar(socioId: string, reservaId: string): Promise<Reserva>;
  reservasDe(socioId: string): Promise<Reserva[]>;
}
