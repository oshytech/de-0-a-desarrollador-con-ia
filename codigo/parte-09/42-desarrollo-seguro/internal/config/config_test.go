package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

const claveDePrueba = "clave-de-mentira-solo-para-tests-0123456789"

func entorno(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestCargar_FallaSiFaltanSecretos(t *testing.T) {
	_, err := Cargar(entorno(map[string]string{}))
	if err == nil {
		t.Fatal("esperaba un error con la configuración vacía")
	}
	for _, esperado := range []string{"RESERVAS_URL_BD", "RESERVAS_CLAVE_TOKENS"} {
		if !strings.Contains(err.Error(), esperado) {
			t.Errorf("el error debería nombrar %s: %v", esperado, err)
		}
	}
}

func TestCargar_ClaveCortaNoApareceEnElError(t *testing.T) {
	_, err := Cargar(entorno(map[string]string{
		"RESERVAS_URL_BD":       "postgres://club:mentira@localhost/club",
		"RESERVAS_CLAVE_TOKENS": "corta-y-secreta",
	}))
	if err == nil {
		t.Fatal("esperaba un error con una clave demasiado corta")
	}
	if strings.Contains(err.Error(), "corta-y-secreta") {
		t.Fatalf("el error filtra el valor del secreto: %v", err)
	}
}

func TestSecreto_NoSeImprimeNiSeRegistra(t *testing.T) {
	cfg, err := Cargar(entorno(map[string]string{
		"RESERVAS_URL_BD":       "postgres://club:mentira@localhost/club",
		"RESERVAS_CLAVE_TOKENS": claveDePrueba,
	}))
	if err != nil {
		t.Fatal(err)
	}

	var salidas []string
	for _, verbo := range []string{"%v", "%+v", "%#v", "%s"} {
		salidas = append(salidas, fmt.Sprintf(verbo, cfg))
	}

	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	log.Info("arranque", "config", cfg, "clave", cfg.ClaveTokens)
	salidas = append(salidas, buf.String())

	j, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	salidas = append(salidas, string(j))

	for _, s := range salidas {
		if strings.Contains(s, claveDePrueba) || strings.Contains(s, "mentira@") {
			t.Errorf("un secreto se ha filtrado en: %s", s)
		}
	}
	if cfg.ClaveTokens.Revelar() != claveDePrueba {
		t.Error("Revelar tiene que devolver el valor real")
	}
}
