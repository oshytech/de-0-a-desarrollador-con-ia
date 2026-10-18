-- 003: el club quiere avisar por SMS. Paso "expandir": la columna nace opcional,
-- así el código que ya está en producción sigue insertando socios sin teléfono.
-- Cuando todos los socios lo tengan, otra migración podrá añadir NOT NULL.
ALTER TABLE socio ADD COLUMN telefono text;
