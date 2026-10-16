-- Carga de prueba y comparación de planes con y sin índice.
-- Uso: docker exec -i club-pg-37 psql -U club -d club < consultas/explicar_indice.sql

-- 20 pistas y 2000 socios de prueba.
INSERT INTO pista (nombre, deporte)
SELECT 'Pista ' || n, CASE WHEN n % 2 = 0 THEN 'padel' ELSE 'tenis' END
FROM generate_series(2, 20) AS n
ON CONFLICT (nombre) DO NOTHING;

INSERT INTO socio (nombre, correo)
SELECT 'Socio ' || n, 'socio' || n || '@ejemplo.org'
FROM generate_series(1, 2000) AS n
ON CONFLICT (correo) DO NOTHING;

-- Un año de reservas: cada pista, cada día, de 08:00 a 21:00.
TRUNCATE reserva;
WITH socios AS (SELECT array_agg(id) AS ids FROM socio),
     horas AS (
         SELECT (d + make_interval(hours => hora)) AT TIME ZONE 'Europe/Madrid' AS inicio
         FROM generate_series(timestamp '2026-01-01', timestamp '2026-12-31', interval '1 day') AS d,
              generate_series(8, 21) AS hora
     )
INSERT INTO reserva (socio_id, pista_id, inicio, fin)
SELECT socios.ids[1 + floor(random() * cardinality(socios.ids))::int],
       pista.id, horas.inicio, horas.inicio + interval '1 hour'
FROM pista, horas, socios;

ANALYZE reserva;
SELECT count(*) AS reservas FROM reserva;

-- La consulta que más se repite: las reservas de un socio a partir de una fecha.
\echo '--- Con el índice reserva_socio_inicio'
EXPLAIN (ANALYZE, COSTS OFF, TIMING OFF, SUMMARY OFF)
SELECT id, pista_id, inicio FROM reserva
WHERE socio_id = 42 AND estado = 'confirmada' AND inicio > timestamptz '2026-11-01'
ORDER BY inicio;

\echo '--- Sin el índice (se borra dentro de una transacción y se deshace)'
BEGIN;
DROP INDEX reserva_socio_inicio;
EXPLAIN (ANALYZE, COSTS OFF, TIMING OFF, SUMMARY OFF)
SELECT id, pista_id, inicio FROM reserva
WHERE socio_id = 42 AND estado = 'confirmada' AND inicio > timestamptz '2026-11-01'
ORDER BY inicio;
ROLLBACK;
