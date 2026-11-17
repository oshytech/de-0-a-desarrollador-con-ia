package club.reservas.dominio

import kotlin.time.Instant

data class Pista(
    val id: String,
    val nombre: String,
    val deporte: String,
)

enum class EstadoReserva { CONFIRMADA, CANCELADA }

data class Reserva(
    val id: String,
    val socioId: String,
    val pistaId: String,
    val inicio: Instant,
    val fin: Instant,
    val estado: EstadoReserva,
    // Hora local del club tal como la envía la API ("10:00"), para mostrarla sin hacer cuentas de zonas horarias.
    val horaLocal: String,
    val fechaLocal: String,
)

/**
 * Una hora libre de una pista. [inicio] es el valor exacto que la API espera al reservar,
 * en hora local del club y sin zona: "2026-11-23T10:00". La app no lo construye: lo devuelve tal cual.
 */
data class HoraLibre(
    val inicio: String,
) {
    val etiqueta: String get() = inicio.substring(11, 16)
}

/** Resultado de hablar con la API. Nunca se lanza una excepción hacia la interfaz. */
sealed interface Resultado<out T> {
    data class Exito<T>(
        val valor: T,
    ) : Resultado<T>

    /** La API ha dicho que no (409, 422, 404...). [mensaje] se puede enseñar al socio. */
    data class Rechazo(
        val codigo: String,
        val mensaje: String,
    ) : Resultado<Nothing>

    /** No se ha podido hablar con la API (sin red, servidor caído, tiempo agotado). */
    data class SinConexion(
        val detalle: String,
    ) : Resultado<Nothing>
}
