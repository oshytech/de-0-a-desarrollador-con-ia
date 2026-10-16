-- 002: la consulta más frecuente es "reservas confirmadas de un socio a partir de ahora".
-- Índice parcial: solo indexa las filas confirmadas, que son las que se consultan.
CREATE INDEX reserva_socio_inicio
    ON reserva (socio_id, inicio)
    WHERE estado = 'confirmada';
