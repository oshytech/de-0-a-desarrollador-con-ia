package dominio

import "time"

// HorasLibres devuelve las franjas de un día que se podrían reservar en este
// momento en una pista, con las mismas reglas que NuevaReserva: horario (R1),
// sin solape con reservas confirmadas (R2) y plazo (R6). Una hora que ya ha
// empezado o que cae a más de 14 días no sale, porque reservarla fallaría.
//
// dia lleva la zona horaria del club: las horas se cuentan en esa zona.
// ocupadas son las reservas de la pista que pueden caer ese día.
func HorasLibres(dia, ahora time.Time, ocupadas []Reserva) []Franja {
	libres := []Franja{}
	for h := PrimeraHora; h <= UltimaHora; h++ {
		franja := Franja{inicio: time.Date(dia.Year(), dia.Month(), dia.Day(), h, 0, 0, 0, dia.Location())}
		inicio := franja.Inicio()
		if !inicio.After(ahora) || inicio.Sub(ahora) > MaxAntelacion {
			continue
		}
		if !seSolapaConAlguna(franja, ocupadas) {
			libres = append(libres, franja)
		}
	}
	return libres
}

func seSolapaConAlguna(franja Franja, reservas []Reserva) bool {
	for _, r := range reservas {
		if r.EstaConfirmada() && r.franja.SeSolapaCon(franja) {
			return true
		}
	}
	return false
}
