package http

import "net/http"

// CabeceraSocio es donde llega la identidad del socio.
//
// En este proyecto la cabecera la escribe quien llama, así que cualquiera
// podría hacerse pasar por otro socio. Es un sitio reservado: en producción
// la rellena un middleware que valida un token (capítulo 44) y la borra si
// viene de fuera. Lo importante es que los manejadores nunca lean la
// identidad del cuerpo de la petición.
const CabeceraSocio = "X-Socio-Id"

func socioAutenticado(w http.ResponseWriter, r *http.Request) (string, bool) {
	socioID := r.Header.Get(CabeceraSocio)
	if socioID == "" {
		escribirJSON(w, http.StatusUnauthorized, cuerpoDeError("no_autenticado", "falta la cabecera "+CabeceraSocio))
		return "", false
	}
	return socioID, true
}
