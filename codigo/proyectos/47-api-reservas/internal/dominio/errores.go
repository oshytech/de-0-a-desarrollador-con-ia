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
	ErrFueraDePlazo = &ReglaIncumplida{
		Codigo:  "fuera_de_plazo",
		Mensaje: "no se reserva en el pasado ni con más de 14 días de antelación",
	}
	ErrLimiteDeReservas = &ReglaIncumplida{
		Codigo:  "limite_de_reservas",
		Mensaje: "un socio no puede tener más de 2 reservas futuras confirmadas",
	}
	ErrPistaOcupada = &ReglaIncumplida{
		Codigo:  "pista_ocupada",
		Mensaje: "la pista ya está reservada a esa hora",
	}

	// Reglas de la cancelación (el reto del capítulo).
	ErrCancelacionFueraDePlazo = &ReglaIncumplida{
		Codigo:  "cancelacion_fuera_de_plazo",
		Mensaje: "solo se puede cancelar hasta 2 horas antes del inicio",
	}
	ErrReservaYaCancelada = &ReglaIncumplida{
		Codigo:  "reserva_ya_cancelada",
		Mensaje: "la reserva ya estaba cancelada",
	}
	ErrReservaDeOtroSocio = &ReglaIncumplida{
		Codigo:  "reserva_de_otro_socio",
		Mensaje: "solo quien hizo la reserva puede cancelarla",
	}
)
