-- 001: socios, pistas y reservas del club.
-- Los identificadores son texto porque los genera la aplicación (puerto NuevoID).
CREATE TABLE socio (
    id     text    PRIMARY KEY,
    nombre text    NOT NULL,
    correo text    NOT NULL UNIQUE,
    activo boolean NOT NULL DEFAULT true
);

CREATE TABLE pista (
    id      text PRIMARY KEY,
    nombre  text NOT NULL UNIQUE,
    deporte text NOT NULL CHECK (deporte IN ('tenis', 'padel'))
);

CREATE TABLE reserva (
    id       text        PRIMARY KEY,
    socio_id text        NOT NULL REFERENCES socio (id),
    pista_id text        NOT NULL REFERENCES pista (id),
    inicio   timestamptz NOT NULL,
    fin      timestamptz NOT NULL,
    estado   text        NOT NULL CHECK (estado IN ('confirmada', 'cancelada')),

    -- Regla 1, también en la base de datos: una hora exacta, en punto,
    -- entre las 08:00 y las 21:00 en hora del club.
    CONSTRAINT reserva_dura_una_hora CHECK (fin = inicio + interval '1 hour'),
    CONSTRAINT reserva_en_punto CHECK (date_trunc('hour', inicio) = inicio),
    CONSTRAINT reserva_en_horario CHECK (
        extract(hour FROM inicio AT TIME ZONE 'Europe/Madrid') BETWEEN 8 AND 21
    )
);

-- Regla 2: como todas las reservas duran una hora y empiezan en punto, dos
-- reservas confirmadas de una pista se solapan solo si empiezan a la vez.
-- Basta un índice único parcial: las canceladas no cuentan y liberan la hora.
CREATE UNIQUE INDEX reserva_pista_inicio_confirmada
    ON reserva (pista_id, inicio)
    WHERE estado = 'confirmada';

-- Para las consultas por socio (regla 3 y listado).
CREATE INDEX reserva_socio_inicio ON reserva (socio_id, inicio);
