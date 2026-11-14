import { ErrorApi, type ApiReservas, type HoraLibre, type Pista, type Reserva } from "./tipos";

// Forma del JSON de la API del capítulo 47. Si cambia, solo se toca este fichero.
interface ReservaJson {
  id: string;
  socio_id: string;
  pista_id: string;
  inicio: string;
  fin: string;
  estado: "confirmada" | "cancelada";
}

interface DisponibilidadJson {
  fecha: string;
  horas_libres: string[]; // ["08:00", "09:00", ...]
}

interface ErrorJson {
  error: string;
  mensaje: string;
}

const MENSAJES_POR_ESTADO: Record<number, string> = {
  404: "No se ha encontrado lo que buscabas.",
  409: "No se ha podido completar: la pista o la reserva han cambiado.",
  422: "La reserva no cumple las normas del club.",
};

export const aReserva = (r: ReservaJson): Reserva => ({
  id: r.id,
  socioId: r.socio_id,
  pistaId: r.pista_id,
  inicio: r.inicio,
  fin: r.fin,
  estado: r.estado,
});

export const aHoraLibre = (fecha: string, hora: string): HoraLibre => ({ inicio: `${fecha}T${hora}`, etiqueta: hora });

/**
 * Cliente HTTP de la API de reservas. Recibe `fetch` como parámetro para poder
 * probarlo sin red: en los tests se le pasa una función que devuelve respuestas fijas.
 */
export function crearApiHttp(urlBase: string, fetchFn: typeof fetch = fetch): ApiReservas {
  async function pedir<T>(ruta: string, opciones: RequestInit = {}): Promise<T> {
    let respuesta: Response;
    try {
      respuesta = await fetchFn(`${urlBase}${ruta}`, opciones);
    } catch {
      throw new ErrorApi(0, "sin_conexion", "No hay conexión con el club. Inténtalo de nuevo en un momento.");
    }
    if (respuesta.ok) return (await respuesta.json()) as T;

    const cuerpo = (await respuesta.json().catch(() => null)) as ErrorJson | null;
    throw new ErrorApi(
      respuesta.status,
      cuerpo?.error ?? `http_${respuesta.status}`,
      cuerpo?.mensaje ?? MENSAJES_POR_ESTADO[respuesta.status] ?? `El servidor ha respondido con un error (${respuesta.status}).`,
    );
  }

  // La API identifica al socio por esta cabecera (en producción la pondría la capa de autenticación).
  const comoSocio = (socioId: string) => ({ "X-Socio-Id": socioId });

  return {
    pistas: () => pedir<Pista[]>("/pistas"),

    async horasLibres(pistaId, fecha) {
      const datos = await pedir<DisponibilidadJson>(
        `/pistas/${encodeURIComponent(pistaId)}/disponibilidad?fecha=${encodeURIComponent(fecha)}`,
      );
      return datos.horas_libres.map((hora) => aHoraLibre(datos.fecha, hora));
    },

    async reservar(socioId, pistaId, inicio) {
      const datos = await pedir<ReservaJson>("/reservas", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...comoSocio(socioId) },
        body: JSON.stringify({ pista_id: pistaId, inicio }),
      });
      return aReserva(datos);
    },

    async cancelar(socioId, reservaId) {
      const datos = await pedir<ReservaJson>(`/reservas/${encodeURIComponent(reservaId)}/cancelacion`, {
        method: "POST",
        headers: comoSocio(socioId),
      });
      return aReserva(datos);
    },

    async reservasDe(socioId) {
      const datos = await pedir<ReservaJson[]>(`/socios/${encodeURIComponent(socioId)}/reservas`, {
        headers: comoSocio(socioId),
      });
      return datos.map(aReserva);
    },
  };
}
