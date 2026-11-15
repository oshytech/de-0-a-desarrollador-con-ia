import { useState } from "react";
import type { ApiReservas } from "./api/tipos";
import { Reservar } from "./features/reservar/Reservar";
import { MisReservas } from "./features/mis-reservas/MisReservas";

interface Props {
  api: ApiReservas;
}

// Sin inicio de sesión: el socio escribe su identificador. Sirve para la demo, no para producción
// (la autenticación de una API se trata en el capítulo 44).
export function App({ api }: Props) {
  const [socioId, setSocioId] = useState("s-1");
  const [version, setVersion] = useState(0);

  return (
    <>
      <header>
        <h1>Reservas del club</h1>
        <label>
          Socio
          <input value={socioId} onChange={(e) => setSocioId(e.target.value.trim())} />
        </label>
      </header>
      <main>
        <Reservar api={api} socioId={socioId} alReservar={() => setVersion((v) => v + 1)} />
        <MisReservas api={api} socioId={socioId} version={version} />
      </main>
    </>
  );
}
