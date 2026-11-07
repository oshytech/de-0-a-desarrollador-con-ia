package dominio

// Socio es una persona del club. Sus datos se pueden leer y escribir
// libremente porque no hay ninguna combinación de valores que rompa una regla.
type Socio struct {
	ID     string
	Nombre string
	Correo string
	Activo bool
}

// PuedeReservar aplica la regla 4: solo reservan los socios activos.
func (s Socio) PuedeReservar() error {
	if !s.Activo {
		return ErrSocioInactivo
	}
	return nil
}
