package reservas

import "time"

// Las interfaces las declara quien las usa, con solo lo que necesita.

type RepositorioReservas interface {
	DePista(pistaID int, dia time.Time) ([]Reserva, error)
	FuturasDeSocio(socioID int, desde time.Time) ([]Reserva, error)
	Guardar(r Reserva) (Reserva, error)
}

type RepositorioSocios interface {
	Buscar(id int) (Socio, bool, error)
}

type Reloj interface {
	Ahora() time.Time
}

type ReservarPista struct {
	reservas RepositorioReservas
	socios   RepositorioSocios
	reloj    Reloj
}

func NuevoReservarPista(reservas RepositorioReservas, socios RepositorioSocios, reloj Reloj) *ReservarPista {
	return &ReservarPista{reservas: reservas, socios: socios, reloj: reloj}
}

func (uc *ReservarPista) Ejecutar(socioID, pistaID int, inicio time.Time) (Reserva, error) {
	ahora := uc.reloj.Ahora()

	socio, existe, err := uc.socios.Buscar(socioID)
	if err != nil {
		return Reserva{}, err
	}
	if !existe {
		return Reserva{}, ErrSocioNoExiste
	}
	if !socio.Activo {
		return Reserva{}, ErrSocioInactivo
	}
	if inicio.Minute() != 0 || inicio.Second() != 0 || inicio.Hour() < Apertura || inicio.Hour() > UltimoInicio {
		return Reserva{}, ErrFueraDeHorario
	}
	if !inicio.After(ahora) || inicio.Sub(ahora) > AntelacionMaxima {
		return Reserva{}, ErrFueraDePlazo
	}

	dePista, err := uc.reservas.DePista(pistaID, inicio)
	if err != nil {
		return Reserva{}, err
	}
	for _, r := range dePista {
		if r.SeSolapaCon(inicio) {
			return Reserva{}, ErrPistaOcupada
		}
	}

	futuras, err := uc.reservas.FuturasDeSocio(socioID, ahora)
	if err != nil {
		return Reserva{}, err
	}
	if len(futuras) >= MaxReservasFuturas {
		return Reserva{}, ErrLimiteReservas
	}

	return uc.reservas.Guardar(Reserva{SocioID: socioID, PistaID: pistaID, Inicio: inicio, Estado: "confirmada"})
}
