package dominio

// ReglaIncumplida es el error que devuelve el dominio cuando una operación
// choca con una regla del club. Codigo es estable y lo usan los adaptadores
// para decidir cómo responder. Mensaje está pensado para personas.
type ReglaIncumplida struct {
	Codigo  string
	Mensaje string
}

func (e *ReglaIncumplida) Error() string { return e.Mensaje }

var (
	ErrFueraDeHorario = &ReglaIncumplida{
		Codigo:  "fuera_de_horario",
		Mensaje: "las reservas empiezan en punto, entre las 08:00 y las 21:00",
	}
	ErrSocioInactivo = &ReglaIncumplida{
		Codigo:  "socio_inactivo",
		Mensaje: "solo los socios activos pueden reservar",
	}
)
