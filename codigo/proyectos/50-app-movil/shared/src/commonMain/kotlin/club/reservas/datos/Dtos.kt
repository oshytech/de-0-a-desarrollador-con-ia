package club.reservas.datos

import club.reservas.dominio.EstadoReserva
import club.reservas.dominio.HoraLibre
import club.reservas.dominio.Pista
import club.reservas.dominio.Reserva
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlin.time.Instant

// Forma exacta del JSON de la API del capítulo 47. Si la API cambia, se toca este fichero y nada más.

@Serializable
data class PistaDto(
    val id: String,
    val nombre: String,
    val deporte: String,
)

@Serializable
data class DisponibilidadDto(
    val fecha: String,
    @SerialName("horas_libres") val horasLibres: List<String>,
)

@Serializable
data class NuevaReservaDto(
    @SerialName("pista_id") val pistaId: String,
    val inicio: String,
)

@Serializable
data class ReservaDto(
    val id: String,
    @SerialName("socio_id") val socioId: String,
    @SerialName("pista_id") val pistaId: String,
    val inicio: String,
    val fin: String,
    val estado: String,
)

@Serializable
data class ErrorDto(
    val error: String,
    val mensaje: String,
)

fun PistaDto.aDominio() = Pista(id = id, nombre = nombre, deporte = deporte)

fun DisponibilidadDto.aDominio() = horasLibres.map { hora -> HoraLibre(inicio = "${fecha}T$hora") }

fun ReservaDto.aDominio() =
    Reserva(
        id = id,
        socioId = socioId,
        pistaId = pistaId,
        inicio = Instant.parse(inicio),
        fin = Instant.parse(fin),
        estado = if (estado == "cancelada") EstadoReserva.CANCELADA else EstadoReserva.CONFIRMADA,
        fechaLocal = inicio.substring(0, 10),
        horaLocal = inicio.substring(11, 16),
    )
