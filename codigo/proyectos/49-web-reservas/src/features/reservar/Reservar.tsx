import { useEffect, useState } from "react";
import { ErrorApi, type ApiReservas, type HoraLibre, type Pista } from "../../api/tipos";
import { diasReservables, nombreDelDia } from "../../dominio/reglas";

type Carga<T> = { tipo: "cargando" } | { tipo: "error"; mensaje: string } | { tipo: "listo"; datos: T };
type Aviso = { tipo: "ok" | "error"; texto: string } | null;

interface Props {
  api: ApiReservas;
  socioId: string;
  ahora?: () => Date;
  alReservar?: () => void;
}

function mensajeDe(error: unknown): string {
  return error instanceof ErrorApi ? error.message : "Algo ha fallado. Inténtalo de nuevo.";
}

export function Reservar({ api, socioId, ahora = () => new Date(), alReservar }: Props) {
  const [dias] = useState(() => diasReservables(ahora()));
  const [pistas, setPistas] = useState<Carga<Pista[]>>({ tipo: "cargando" });
  const [pistaId, setPistaId] = useState<string | null>(null);
  const [fecha, setFecha] = useState(dias[0]);
  const [horas, setHoras] = useState<Carga<HoraLibre[]> | null>(null);
  const [recargas, setRecargas] = useState(0);
  const [enviando, setEnviando] = useState<string | null>(null);
  const [aviso, setAviso] = useState<Aviso>(null);

  useEffect(() => {
    api
      .pistas()
      .then((datos) => setPistas({ tipo: "listo", datos }))
      .catch((e) => setPistas({ tipo: "error", mensaje: mensajeDe(e) }));
  }, [api]);

  useEffect(() => {
    if (pistaId === null) return;
    let vigente = true; // evita pintar una respuesta vieja si el socio cambia rápido de día
    setHoras({ tipo: "cargando" });
    api
      .horasLibres(pistaId, fecha)
      .then((datos) => vigente && setHoras({ tipo: "listo", datos }))
      .catch((e) => vigente && setHoras({ tipo: "error", mensaje: mensajeDe(e) }));
    return () => {
      vigente = false;
    };
  }, [api, pistaId, fecha, recargas]);

  async function reservar(hora: HoraLibre) {
    if (pistaId === null) return;
    setEnviando(hora.inicio);
    setAviso(null);
    try {
      await api.reservar(socioId, pistaId, hora.inicio);
      const pista = pistas.tipo === "listo" ? pistas.datos.find((p) => p.id === pistaId) : undefined;
      setAviso({
        tipo: "ok",
        texto: `Reserva confirmada: ${pista?.nombre ?? "pista"}, ${nombreDelDia(fecha)} a las ${hora.etiqueta}.`,
      });
      alReservar?.();
    } catch (e) {
      setAviso({ tipo: "error", texto: mensajeDe(e) });
    } finally {
      setEnviando(null);
      setRecargas((n) => n + 1); // la hora reservada (o la que otro nos quitó) desaparece de la lista
    }
  }

  if (pistas.tipo === "cargando") return <p>Cargando pistas…</p>;
  if (pistas.tipo === "error") return <p role="alert">{pistas.mensaje}</p>;

  return (
    <section aria-labelledby="titulo-reservar">
      <h2 id="titulo-reservar">Reservar pista</h2>

      <div className="campos">
        <label>
          Pista
          <select
            value={pistaId ?? ""}
            onChange={(e) => setPistaId(e.target.value || null)}
          >
            <option value="">Elige una pista</option>
            {pistas.datos.map((p) => (
              <option key={p.id} value={p.id}>
                {p.nombre} ({p.deporte === "padel" ? "pádel" : "tenis"})
              </option>
            ))}
          </select>
        </label>

        <label>
          Día
          <select value={fecha} onChange={(e) => setFecha(e.target.value)}>
            {dias.map((d) => (
              <option key={d} value={d}>
                {nombreDelDia(d)}
              </option>
            ))}
          </select>
        </label>
      </div>

      <p role="status" className={aviso?.tipo === "ok" ? "aviso-ok" : undefined}>
        {aviso?.tipo === "ok" ? aviso.texto : ""}
      </p>
      {aviso?.tipo === "error" && (
        <p role="alert" className="aviso-error">
          {aviso.texto}
        </p>
      )}

      {pistaId === null && <p>Elige una pista para ver las horas libres.</p>}
      {horas?.tipo === "cargando" && <p>Cargando horas libres…</p>}
      {horas?.tipo === "error" && <p role="alert">{horas.mensaje}</p>}
      {horas?.tipo === "listo" && horas.datos.length === 0 && <p>No quedan horas libres ese día en esta pista.</p>}
      {horas?.tipo === "listo" && horas.datos.length > 0 && (
        <ul className="horas" aria-label="Horas libres">
          {horas.datos.map((h) => (
            <li key={h.inicio}>
              <button
                type="button"
                onClick={() => reservar(h)}
                disabled={enviando !== null}
                aria-label={`Reservar a las ${h.etiqueta}`}
              >
                {enviando === h.inicio ? "Reservando…" : h.etiqueta}
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
