-- Datos de ejemplo para probar la API en local, los mismos que
-- memoria.DatosDeEjemplo. Se cargan solo con DATOS_DE_EJEMPLO=si.
INSERT INTO socio (id, nombre, correo, activo) VALUES
    ('s-1', 'Ana', 'ana@example.com', true),
    ('s-2', 'Luis', 'luis@example.com', false),
    ('s-3', 'Marta', 'marta@example.com', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO pista (id, nombre, deporte) VALUES
    ('p-1', 'Pista 1', 'tenis'),
    ('p-2', 'Pista 2', 'padel')
ON CONFLICT (id) DO NOTHING;
