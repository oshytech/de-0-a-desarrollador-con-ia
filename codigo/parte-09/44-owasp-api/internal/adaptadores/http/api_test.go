package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"owaspapi/internal/dominio"
)

const (
	ana   int64 = 1
	bruno int64 = 2
)

func socios() *Socios {
	return &Socios{Datos: map[int64]dominio.Socio{
		ana:   {ID: ana, Nombre: "Ana", Correo: "ana@example.com", Activo: true},
		bruno: {ID: bruno, Nombre: "Bruno", Correo: "bruno@example.com", Activo: false, NotasInternas: "cuota de octubre pendiente"},
	}}
}

var sesiones = Sesiones{"token-ana": ana, "token-bruno": bruno}

func patch(h http.Handler, token, cuerpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPatch, "/socios/yo", strings.NewReader(cuerpo))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// --- API3: autorización a nivel de propiedad ---

func TestPerfilVulnerable_BrunoSeReactivaYVeSusNotas(t *testing.T) {
	s := socios()
	h := sesiones.Identificar(ActualizarPerfilVulnerable(s))

	rec := patch(h, "token-bruno", `{"nombre":"Bruno","activo":true}`)

	if !s.leer(bruno).Activo {
		t.Fatal("esperaba demostrar el mass assignment: Bruno debería haberse activado")
	}
	if !strings.Contains(rec.Body.String(), "cuota de octubre") {
		t.Fatal("esperaba demostrar la exposición excesiva de notas internas")
	}
}

func TestPerfil_SoloAceptaLosCamposPermitidos(t *testing.T) {
	s := socios()
	h := sesiones.Identificar(ActualizarPerfil(s))

	for _, cuerpo := range []string{
		`{"activo":true}`,
		`{"nombre":"Bruno","activo":true}`,
		`{"notas_internas":""}`,
		`{"id":1,"nombre":"Bruno"}`,
	} {
		if rec := patch(h, "token-bruno", cuerpo); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: estado %d, esperaba 400", cuerpo, rec.Code)
		}
	}
	b := s.leer(bruno)
	if b.Activo || b.NotasInternas == "" {
		t.Fatalf("los campos de administración no deberían haber cambiado: %+v", b)
	}
	if s.leer(ana).Nombre != "Ana" {
		t.Fatal("el perfil de Ana no debería haber cambiado")
	}
}

func TestPerfil_CambioLegitimoYRespuestaMinima(t *testing.T) {
	s := socios()
	h := sesiones.Identificar(ActualizarPerfil(s))

	rec := patch(h, "token-bruno", `{"correo":"bruno.garcia@example.com"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado %d: %s", rec.Code, rec.Body.String())
	}
	var cuerpo map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &cuerpo); err != nil {
		t.Fatal(err)
	}
	for _, campo := range []string{"id", "notas_internas"} {
		if _, sale := cuerpo[campo]; sale {
			t.Errorf("la respuesta no debería incluir %q", campo)
		}
	}
	if s.leer(bruno).Correo != "bruno.garcia@example.com" {
		t.Fatal("el correo debería haber cambiado")
	}
	if rec := patch(h, "token-bruno", `{"correo":"esto no es un correo"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("un correo no válido debería dar 400, dio %d", rec.Code)
	}
}

// --- API4: consumo de recursos sin restringir ---

type relojManual struct{ t time.Time }

func (r *relojManual) Ahora() time.Time        { return r.t }
func (r *relojManual) Avanzar(d time.Duration) { r.t = r.t.Add(d) }

func reservar() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) })
}

func post(h http.Handler, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/reservas", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestLimitador_CortaLaRafagaYSeRecupera(t *testing.T) {
	reloj := &relojManual{t: time.Date(2026, 11, 30, 10, 0, 0, 0, time.UTC)}
	lim := &Limitador{Rafaga: 5, Intervalo: 12 * time.Second, Ahora: reloj.Ahora, Clave: PorSocio}
	h := sesiones.Identificar(lim.Middleware(reservar()))

	for i := 1; i <= 5; i++ {
		if rec := post(h, "token-ana"); rec.Code != http.StatusCreated {
			t.Fatalf("petición %d: estado %d, esperaba 201", i, rec.Code)
		}
	}
	rec := post(h, "token-ana")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("la sexta petición debería dar 429, dio %d", rec.Code)
	}
	if s, _ := strconv.Atoi(rec.Header().Get("Retry-After")); s != 12 {
		t.Fatalf("Retry-After = %q, esperaba 12", rec.Header().Get("Retry-After"))
	}

	// Bruno tiene su propio cubo: lo que hace Ana no le afecta.
	if rec := post(h, "token-bruno"); rec.Code != http.StatusCreated {
		t.Fatalf("Bruno no debería estar limitado, estado %d", rec.Code)
	}

	reloj.Avanzar(12 * time.Second)
	if rec := post(h, "token-ana"); rec.Code != http.StatusCreated {
		t.Fatalf("tras 12 s Ana debería tener una ficha, estado %d", rec.Code)
	}
	if rec := post(h, "token-ana"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("pero solo una, estado %d", rec.Code)
	}
}

func TestLimitador_NoAcumulaClavesSinLimite(t *testing.T) {
	reloj := &relojManual{t: time.Date(2026, 11, 30, 10, 0, 0, 0, time.UTC)}
	lim := &Limitador{Rafaga: 5, Intervalo: time.Second, Ahora: reloj.Ahora}

	for i := 0; i < 10_000; i++ {
		lim.Permitir("ip:" + strconv.Itoa(i))
	}
	reloj.Avanzar(time.Minute)
	lim.Permitir("ip:nueva")

	if n := lim.numeroDeCubos(); n != 1 {
		t.Fatalf("tras la purga debería quedar 1 cubo, quedan %d", n)
	}
}
