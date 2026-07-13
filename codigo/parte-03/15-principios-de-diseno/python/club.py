"""Después: cada regla es una función, y reservar() solo coordina."""
from dataclasses import dataclass
from datetime import datetime, timedelta

# Cada constante es una pieza de conocimiento del club, escrita una sola vez.
APERTURA = 8
ULTIMO_INICIO = 21
DURACION = timedelta(hours=1)
ANTELACION_MAXIMA = timedelta(days=14)
MAX_RESERVAS_FUTURAS = 2
MARGEN_CANCELACION = timedelta(hours=2)  # vale 2, como la anterior, pero es otra regla


@dataclass(frozen=True)
class Socio:
    id: int
    nombre: str
    correo: str
    activo: bool


@dataclass
class Reserva:
    id: int
    socio: Socio
    pista: str
    inicio: datetime
    estado: str = "confirmada"

    @property
    def fin(self):
        return self.inicio + DURACION


@dataclass(frozen=True)
class Solicitud:
    socio: Socio
    pista: str
    inicio: datetime


class ReservaRechazada(Exception):
    pass


# --- Reglas: reciben la solicitud y el contexto, y devuelven un motivo o None ---

def socio_activo(s, reservas, ahora):
    if not s.socio.activo:
        return "el socio no está activo"


def en_horario(s, reservas, ahora):
    en_punto = s.inicio.minute == 0 and s.inicio.second == 0
    if not en_punto or not APERTURA <= s.inicio.hour <= ULTIMO_INICIO:
        return "las reservas empiezan en punto, entre las 08:00 y las 21:00"


def dentro_de_plazo(s, reservas, ahora):
    if s.inicio <= ahora:
        return "no se puede reservar en el pasado"
    if s.inicio - ahora > ANTELACION_MAXIMA:
        return "no se puede reservar con más de 14 días de antelación"


def pista_libre(s, reservas, ahora):
    for r in reservas:
        if r.pista == s.pista and r.estado == "confirmada" and r.inicio < s.inicio + DURACION and s.inicio < r.fin:
            return "la pista ya está reservada a esa hora"


def bajo_el_limite(s, reservas, ahora):
    futuras = [r for r in reservas
               if r.socio.id == s.socio.id and r.estado == "confirmada" and r.inicio > ahora]
    if len(futuras) >= MAX_RESERVAS_FUTURAS:
        return "ya tienes 2 reservas pendientes"


REGLAS = [socio_activo, en_horario, dentro_de_plazo, pista_libre, bajo_el_limite]


# --- Caso de uso: coordina, pero no sabe cómo se guarda ni cómo se avisa ---

class RepositorioEnMemoria:
    def __init__(self):
        self._reservas = []

    def todas(self):
        return list(self._reservas)

    def guardar(self, reserva):
        self._reservas.append(reserva)

    def siguiente_id(self):
        return len(self._reservas) + 1


def reservar(solicitud, repositorio, notificar, ahora, reglas=REGLAS):
    reservas = repositorio.todas()
    motivos = [m for regla in reglas if (m := regla(solicitud, reservas, ahora))]
    if motivos:
        raise ReservaRechazada(motivos)
    reserva = Reserva(repositorio.siguiente_id(), solicitud.socio, solicitud.pista, solicitud.inicio)
    repositorio.guardar(reserva)
    notificar(solicitud.socio, f"Reserva confirmada: {reserva.pista}, {reserva.inicio:%d/%m %H:%M}")
    return reserva


def notificar_por_consola(socio, mensaje):
    print(f"[aviso a {socio.correo}] {mensaje}")


if __name__ == "__main__":
    import sys
    sys.stdout.reconfigure(encoding="utf-8")
    repo = RepositorioEnMemoria()
    lucia = Socio(1, "Lucía", "lucia@example.com", activo=True)
    ahora = datetime(2026, 3, 10, 12, 0)
    reservar(Solicitud(lucia, "Pádel 1", datetime(2026, 3, 11, 19, 0)), repo, notificar_por_consola, ahora)
    try:
        reservar(Solicitud(lucia, "Pádel 1", datetime(2026, 3, 11, 19, 30)), repo, notificar_por_consola, ahora)
    except ReservaRechazada as e:
        print("Rechazada:", " | ".join(e.args[0]))
