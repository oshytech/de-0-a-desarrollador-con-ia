-- 001: socios, pistas y reservas del club.
-- btree_gist permite mezclar "=" sobre un número y "&&" sobre un rango
-- dentro de la misma restricción de exclusión.
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE socio (
    id     bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nombre text    NOT NULL,
    correo text    NOT NULL UNIQUE,
    activo boolean NOT NULL DEFAULT true
);

CREATE TABLE pista (
    id      bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nombre  text NOT NULL UNIQUE,
    deporte text NOT NULL CHECK (deporte IN ('tenis', 'padel'))
);

CREATE TABLE reserva (
    id       bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    socio_id bigint      NOT NULL REFERENCES socio (id),
    pista_id bigint      NOT NULL REFERENCES pista (id),
    inicio   timestamptz NOT NULL,
    fin      timestamptz NOT NULL,
    estado   text        NOT NULL DEFAULT 'confirmada'
             CHECK (estado IN ('confirmada', 'cancelada')),

    -- Regla 1: una hora exacta, en punto, entre las 08:00 y las 21:00 (hora del club).
    CONSTRAINT reserva_dura_una_hora CHECK (fin = inicio + interval '1 hour'),
    CONSTRAINT reserva_en_punto CHECK (date_trunc('hour', inicio) = inicio),
    CONSTRAINT reserva_en_horario CHECK (
        extract(hour FROM inicio AT TIME ZONE 'Europe/Madrid') BETWEEN 8 AND 21
    ),

    -- Regla 2: una pista no puede tener dos reservas confirmadas que se solapen.
    CONSTRAINT reserva_sin_solape EXCLUDE USING gist (
        pista_id WITH =,
        tstzrange(inicio, fin) WITH &&
    ) WHERE (estado = 'confirmada')
);
