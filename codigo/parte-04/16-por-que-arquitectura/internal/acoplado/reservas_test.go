package acoplado

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Para comprobar una regla de fechas hay que montar una petición HTTP,
// escribir JSON a mano y calcular la fecha a partir de la hora real.
func TestRechazaReservasConMasDe14Dias(t *testing.T) {
	dentroDe20Dias := time.Now().AddDate(0, 0, 20).Format("2006-01-02") + "T10:00"
	cuerpo := `{"socio":"ana","pista":"1","inicio":"` + dentroDe20Dias + `"}`
	respuesta := httptest.NewRecorder()

	ManejarReserva(respuesta, httptest.NewRequest("POST", "/reservas", strings.NewReader(cuerpo)))

	if respuesta.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperaba 422, salió %d", respuesta.Code)
	}
}
