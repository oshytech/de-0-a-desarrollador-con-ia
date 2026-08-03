package aplicacion

import (
	"context"
	"fmt"
	"sync"

	"reservas/internal/dominio"
)

// AvisarAlSocio reacciona a los eventos de reserva y escribe al socio.
// Es idempotente: si el mismo evento llega dos veces, solo avisa una.
type AvisarAlSocio struct {
	socios      RepositorioSocios
	notificador Notificador

	mu         sync.Mutex
	procesados map[string]bool // en producción viviría en la base de datos
}

func NuevoAvisarAlSocio(socios RepositorioSocios, notificador Notificador) *AvisarAlSocio {
	return &AvisarAlSocio{socios: socios, notificador: notificador, procesados: map[string]bool{}}
}

func (uc *AvisarAlSocio) Manejar(ctx context.Context, evento dominio.Evento) error {
	if uc.yaProcesado(evento.Clave()) {
		return nil
	}
	var socioID, asunto, texto string
	switch e := evento.(type) {
	case dominio.ReservaConfirmada:
		socioID, asunto = e.SocioID, "Reserva confirmada"
		texto = fmt.Sprintf("Tienes la pista %s el %s.", e.PistaID, e.Inicio.Format("02/01 a las 15:04"))
	case dominio.ReservaCancelada:
		socioID, asunto = e.SocioID, "Reserva cancelada"
		texto = fmt.Sprintf("Has cancelado la pista %s del %s.", e.PistaID, e.Inicio.Format("02/01 a las 15:04"))
	default:
		return nil // los eventos que no le interesan se ignoran
	}
	socio, err := uc.socios.BuscarPorID(ctx, socioID)
	if err != nil {
		return fmt.Errorf("socio %q: %w", socioID, err)
	}
	if err := uc.notificador.Enviar(ctx, socio.Correo, asunto, texto); err != nil {
		return fmt.Errorf("avisar a %q: %w", socio.Correo, err)
	}
	uc.marcarProcesado(evento.Clave())
	return nil
}

func (uc *AvisarAlSocio) yaProcesado(clave string) bool {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.procesados[clave]
}

func (uc *AvisarAlSocio) marcarProcesado(clave string) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.procesados[clave] = true
}
