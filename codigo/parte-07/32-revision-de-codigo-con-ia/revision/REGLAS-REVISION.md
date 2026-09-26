# Reglas de revisión del proyecto de reservas

Revisa solo el diff que se te pasa. Audita contra estas reglas y nada más.
Si algo te parece mal pero no lo cubre ninguna regla, ponlo aparte en
"Fuera de reglas", sin severidad.

## Reglas

1. **El dominio solo importa la librería estándar.**
   Porque `internal/dominio` tiene que poder probarse y reutilizarse sin base de datos ni HTTP.
2. **Nadie llama a `time.Now()` fuera de `cmd/` y `internal/adaptadores/`.** La hora entra por el puerto `Reloj`.
   Porque las reglas de plazo (cancelar hasta 2 horas antes, reservar como mucho a 14 días) dependen de la hora y los tests tienen que poder fijarla.
3. **Un caso de uso que recibe un `socioID` comprueba que la reserva es de ese socio** antes de modificarla.
   Porque si no, cualquier socio puede cambiar las reservas de otro.
4. **Los errores de dominio se comparan con `errors.Is`**, nunca por el texto del mensaje.
   Porque los mensajes se cambian para el usuario y la comparación por texto se rompe sin avisar.
5. **Todo cambio de comportamiento llega con un test que falla sin el cambio.**
   Porque un test que nunca has visto fallar no demuestra nada.

## Formato de cada hallazgo

    [SEVERIDAD] Regla N · fichero:línea
    Qué pasa: una frase.
    Cómo reproducirlo: entrada y estado concretos, o "NO VERIFICADO".

Severidades: BLOQUEANTE, ALTA, MEDIA, BAJA. Sin elogios ni resumen final.
No comentes formato ni orden de imports: eso lo resuelve `gofmt` en la puerta de calidad.
