package club.reservas.datos

import club.reservas.aplicacion.DIAS_DE_ANTELACION
import club.reservas.aplicacion.MARGEN_CANCELACION
import club.reservas.aplicacion.Reloj
import club.reservas.aplicacion.RepositorioReservas
import club.reservas.dominio.EstadoReserva
import club.reservas.dominio.HoraLibre
import club.reservas.dominio.Pista
import club.reservas.dominio.Reserva
import club.reservas.dominio.Resultado
import kotlinx.datetime.DateTimeUnit
import kotlinx.datetime.LocalDate
import kotlinx.datetime.LocalDateTime
import kotlinx.datetime.plus
import kotlinx.datetime.toInstant
import kotlinx.datetime.toLocalDateTime
import kotlin.time.Duration.Companion.hours

/**
 * API de mentira en memoria, con las reglas del club y los mismos datos de ejemplo que la
 * API del capítulo 47. Sirve para desarrollar y probar la app sin levantar el servidor.
 */
class RepositorioSimulado(
    private val reloj: Reloj,
) : RepositorioReservas {
    private val pistas = listOf(Pista("p-1", "Pista 1", "tenis"), Pista("p-2", "Pista 2", "padel"))
    private val sociosInactivos = setOf("s-2")
    private val reservas = mutableListOf<Reserva>()
    private var siguienteId = 1

    override suspend fun pistas(): Resultado<List<Pista>> = Resultado.Exito(pistas)

    override suspend fun horasLibres(
        pistaId: String,
        fecha: LocalDate,
    ): Resultado<List<HoraLibre>> {
        if (pistas.none { it.id == pistaId }) return Resultado.Rechazo("no_encontrado", "Esa pista no existe.")
        val libres =
            (8..21)
                .map { hora -> HoraLibre("${fecha}T${hora.toString().padStart(2, '0')}:00") }
                .filter { hora ->
                    val momento = LocalDateTime.parse(hora.inicio).toInstant(reloj.zona)
                    momento > reloj.ahora() && ocupada(pistaId, hora.inicio) == null
                }
        return Resultado.Exito(libres)
    }

    override suspend fun reservar(
        socioId: String,
        pistaId: String,
        inicio: String,
    ): Resultado<Reserva> {
        val local = LocalDateTime.parse(inicio)
        val momento = local.toInstant(reloj.zona)
        val ahora = reloj.ahora()
        val hoy = ahora.toLocalDateTime(reloj.zona).date
        val pendientes = reservas.count { it.socioId == socioId && it.estado == EstadoReserva.CONFIRMADA && it.inicio > ahora }
        return when {
            socioId in sociosInactivos ->
                Resultado.Rechazo("socio_inactivo", "Tu cuenta no está activa. Pasa por recepción.")
            local.minute != 0 || local.hour !in 8..21 ->
                Resultado.Rechazo("horario_no_valido", "Las reservas empiezan en punto, de 08:00 a 21:00.")
            momento <= ahora || local.date > hoy.plus(DIAS_DE_ANTELACION, DateTimeUnit.DAY) ->
                Resultado.Rechazo("fuera_de_plazo", "Solo se puede reservar desde ahora hasta dentro de 14 días.")
            ocupada(pistaId, inicio) != null ->
                Resultado.Rechazo("pista_ocupada", "Esa hora acaba de reservarla otra persona.")
            pendientes >= 2 ->
                Resultado.Rechazo("limite_de_reservas", "Ya tienes 2 reservas pendientes. Cancela una para reservar otra.")
            else -> {
                val reserva =
                    Reserva(
                        id = "r-${siguienteId++}",
                        socioId = socioId,
                        pistaId = pistaId,
                        inicio = momento,
                        fin = momento + 1.hours,
                        estado = EstadoReserva.CONFIRMADA,
                        fechaLocal = inicio.substring(0, 10),
                        horaLocal = inicio.substring(11, 16),
                    )
                reservas += reserva
                Resultado.Exito(reserva)
            }
        }
    }

    override suspend fun cancelar(
        socioId: String,
        reservaId: String,
    ): Resultado<Reserva> {
        // La reserva de otro socio se responde como si no existiera, igual que en la API real.
        val reserva =
            reservas.find { it.id == reservaId && it.socioId == socioId }
                ?: return Resultado.Rechazo("no_encontrado", "Esa reserva no existe.")
        if (reserva.estado == EstadoReserva.CANCELADA) {
            return Resultado.Rechazo("reserva_ya_cancelada", "Esa reserva ya estaba cancelada.")
        }
        if (reserva.inicio - reloj.ahora() < MARGEN_CANCELACION) {
            return Resultado.Rechazo("cancelacion_fuera_de_plazo", "Solo se puede cancelar hasta 2 horas antes del inicio.")
        }
        val cancelada = reserva.copy(estado = EstadoReserva.CANCELADA)
        reservas[reservas.indexOf(reserva)] = cancelada
        return Resultado.Exito(cancelada)
    }

    override suspend fun reservasDe(socioId: String): Resultado<List<Reserva>> = Resultado.Exito(reservas.filter { it.socioId == socioId })

    private fun ocupada(
        pistaId: String,
        inicioLocal: String,
    ) = reservas.find {
        it.pistaId == pistaId && it.estado == EstadoReserva.CONFIRMADA && "${it.fechaLocal}T${it.horaLocal}" == inicioLocal
    }
}
