package dominio

import (
	"errors"
	"time"
)

// Errores de dominio. Los casos de uso los devuelven tal cual y el adaptador
// HTTP los traduce a códigos de respuesta, así que conviene compararlos con
// errors.Is y no por su texto.
var (
	ErrHorarioNoValido   = errors.New("la reserva debe empezar en punto entre las 08:00 y las 21:00")
	ErrPistaOcupada      = errors.New("la pista ya tiene una reserva confirmada a esa hora")
	ErrLimiteReservas    = errors.New("el socio ya tiene 2 reservas confirmadas futuras")
	ErrSocioInactivo     = errors.New("el socio no está activo")
	ErrFueraDePlazo      = errors.New("no se reserva en el pasado ni con más de 14 días de antelación")
	ErrCancelacionTardia = errors.New("solo se puede cancelar hasta 2 horas antes del inicio")
)

const (
	// DuracionReserva es lo que dura cualquier reserva. No hay medias horas.
	DuracionReserva = time.Hour
	// AntelacionMaxima es cuánto se puede adelantar una reserva.
	AntelacionMaxima = 14 * 24 * time.Hour
	// MargenCancelacion es el tiempo mínimo antes del inicio para cancelar.
	MargenCancelacion = 2 * time.Hour

	primeraHora = 8
	ultimaHora  = 21
)

// Socio es una persona del club. Solo los socios activos pueden reservar.
type Socio struct {
	ID     string
	Nombre string
	Correo string
	Activo bool
}

// Pista es una pista reservable. Deporte vale "tenis" o "pádel".
type Pista struct {
	ID      string
	Nombre  string
	Deporte string
}

// EstadoReserva indica si una reserva cuenta para los solapes y los límites.
type EstadoReserva string

// Estados posibles de una reserva. Solo las confirmadas ocupan pista.
const (
	Confirmada EstadoReserva = "confirmada"
	Cancelada  EstadoReserva = "cancelada"
)

// Reserva es el alquiler de una pista por un socio durante una hora.
// Inicio y Fin se guardan en la zona horaria del club.
type Reserva struct {
	ID      string
	SocioID string
	PistaID string
	Inicio  time.Time
	Fin     time.Time
	Estado  EstadoReserva
}

// PuedeReservar devuelve ErrSocioInactivo si el socio está dado de baja.
func (s Socio) PuedeReservar() error {
	if !s.Activo {
		return ErrSocioInactivo
	}
	return nil
}

// NuevaReserva crea una reserva confirmada de una hora a partir de inicio.
//
// Comprueba el horario (en punto, de 08:00 a 21:00) y el plazo (ni en el
// pasado ni a más de 14 días de ahora). No comprueba solapes ni el límite
// por socio, porque para eso hay que mirar otras reservas.
func NuevaReserva(id, socioID, pistaID string, inicio, ahora time.Time) (Reserva, error) {
	if inicio.Minute() != 0 || inicio.Second() != 0 || inicio.Nanosecond() != 0 ||
		inicio.Hour() < primeraHora || inicio.Hour() > ultimaHora {
		return Reserva{}, ErrHorarioNoValido
	}
	if !inicio.After(ahora) || inicio.Sub(ahora) > AntelacionMaxima {
		return Reserva{}, ErrFueraDePlazo
	}
	return Reserva{
		ID:      id,
		SocioID: socioID,
		PistaID: pistaID,
		Inicio:  inicio,
		Fin:     inicio.Add(DuracionReserva),
		Estado:  Confirmada,
	}, nil
}

// SeSolapaCon indica si dos reservas confirmadas de la misma pista coinciden
// en el tiempo. Una reserva que termina a las 19:00 no se solapa con otra que
// empieza a las 19:00.
func (r Reserva) SeSolapaCon(otra Reserva) bool {
	if r.PistaID != otra.PistaID || r.Estado != Confirmada || otra.Estado != Confirmada {
		return false
	}
	return r.Inicio.Before(otra.Fin) && otra.Inicio.Before(r.Fin)
}

// Cancelar marca la reserva como cancelada si faltan al menos 2 horas para
// el inicio. Cancelar una reserva ya cancelada no hace nada.
func (r *Reserva) Cancelar(ahora time.Time) error {
	if r.Estado == Cancelada {
		return nil
	}
	if r.Inicio.Sub(ahora) < MargenCancelacion {
		return ErrCancelacionTardia
	}
	r.Estado = Cancelada
	return nil
}
