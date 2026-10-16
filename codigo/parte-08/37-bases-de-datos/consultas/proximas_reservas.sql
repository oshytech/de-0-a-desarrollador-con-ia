-- Próximas reservas confirmadas, con el nombre del socio y de la pista.
SELECT s.nombre, p.nombre AS pista, r.inicio
FROM reserva AS r
JOIN socio AS s ON s.id = r.socio_id
JOIN pista AS p ON p.id = r.pista_id
WHERE r.estado = 'confirmada' AND r.inicio > now()
ORDER BY r.inicio;
