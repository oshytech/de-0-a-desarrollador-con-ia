import { useCallback, useEffect, useState } from "react";
import { ErrorApi, type ApiReservas, type Reserva } from "../../api/tipos";
import { nombreDelDia, sePuedeCancelar } from "../../dominio/reglas";

interface Props {
  api: ApiReservas;
  socioId: string;
  ahora?: () => Date;
  /** Cambia cada vez que se hace una reserva en otra parte de la página, para recargar la lista. */
  version?: number;
}

const relojDelSistema = () => new Date();

export function MisReservas({ api, socioId, ahora = relojDelSistema, version = 0 }: Props) {
  const [reservas, setReservas] = useState<Reserva[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [aviso, setAviso] = useState("");
  const [nombres, setNombres] = useState<Record<string, string>>({});

  useEffect(() => {
    api
      .pistas()
      .then((pistas) => setNombres(Object.fromEntries(pistas.map((p) => [p.id, p.nombre]))))
      .catch(() => setNombres({})); // sin nombres se enseña el identificador, no es grave
  }, [api]);

  const cargar = useCallback(async () => {
    try {
      const todas = await api.reservasDe(socioId);
      const momento = ahora().getTime();
      setReservas(
        todas
          .filter((r) => r.estado === "confirmada" && Date.parse(r.inicio) > momento)
          .sort((a, b) => Date.parse(a.inicio) - Date.parse(b.inicio)),
      );
      setError(null);
    } catch (e) {
      setError(e instanceof ErrorApi ? e.message : "No se han podido cargar tus reservas.");
    }
  }, [api, socioId, ahora]);

  useEffect(() => {
    void cargar();
  }, [cargar, version]);

  async function cancelar(reserva: Reserva) {
    try {
      await api.cancelar(socioId, reserva.id);
      setAviso("Reserva cancelada.");
    } catch (e) {
      setError(e instanceof ErrorApi ? e.message : "No se ha podido cancelar.");
    }
    await cargar();
  }

  return (
    <section aria-labelledby="titulo-mis-reservas">
      <h2 id="titulo-mis-reservas">Mis reservas</h2>
      <p role="status">{aviso}</p>
      {error && <p role="alert">{error}</p>}
      {reservas === null && !error && <p>Cargando tus reservas…</p>}
      {reservas?.length === 0 && <p>No tienes reservas pendientes.</p>}
      {reservas && reservas.length > 0 && (
        <ul className="reservas">
          {reservas.map((r) => {
            const cuando = `${nombreDelDia(r.inicio.slice(0, 10))} a las ${r.inicio.slice(11, 16)}`;
            const sePuede = sePuedeCancelar(r, ahora());
            return (
              <li key={r.id}>
                <span>
                  {nombres[r.pistaId] ?? r.pistaId}, {cuando}
                </span>
                {sePuede ? (
                  <button type="button" onClick={() => cancelar(r)} aria-label={`Cancelar la reserva del ${cuando}`}>
                    Cancelar
                  </button>
                ) : (
                  <small>Ya no se puede cancelar (faltan menos de 2 horas).</small>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
