package club.reservas.aplicacion

import club.reservas.dominio.EstadoReserva
import club.reservas.dominio.HoraLibre
import club.reservas.dominio.Pista
import club.reservas.dominio.Reserva
import club.reservas.dominio.Resultado
import kotlinx.datetime.DateTimeUnit
import kotlinx.datetime.LocalDate
import kotlinx.datetime.plus
import kotlinx.datetime.toLocalDateTime
import kotlin.time.Duration.Companion.hours

const val DIAS_DE_ANTELACION = 14
val MARGEN_CANCELACION = 2.hours

/** Regla 6: se reserva desde hoy hasta dentro de 14 días. La app solo ofrece esas fechas. */
class DiasReservables(
    private val reloj: Reloj,
) {
    operator fun invoke(): List<LocalDate> {
        val hoy = reloj.ahora().toLocalDateTime(reloj.zona).date
        return (0..DIAS_DE_ANTELACION).map { hoy.plus(it, DateTimeUnit.DAY) }
    }
}

class ListarPistas(
    private val repositorio: RepositorioReservas,
) {
    suspend operator fun invoke(): Resultado<List<Pista>> = repositorio.pistas()
}

class ConsultarHorasLibres(
    private val repositorio: RepositorioReservas,
    private val diasReservables: DiasReservables,
) {
    suspend operator fun invoke(
        pistaId: String,
        fecha: LocalDate,
    ): Resultado<List<HoraLibre>> {
        if (fecha !in diasReservables()) {
            return Resultado.Rechazo("fuera_de_plazo", "Solo se puede reservar desde hoy hasta dentro de 14 días.")
        }
        return repositorio.horasLibres(pistaId, fecha)
    }
}

class ReservarPista(
    private val repositorio: RepositorioReservas,
) {
    suspend operator fun invoke(
        socioId: String,
        pistaId: String,
        hora: HoraLibre,
    ): Resultado<Reserva> = repositorio.reservar(socioId, pistaId, hora.inicio)
}

/**
 * Regla 5: se puede cancelar hasta 2 horas antes. La API es quien decide, pero comprobarlo
 * aquí permite desactivar el botón y ahorrar una llamada que ya sabemos que va a fallar.
 */
class SePuedeCancelar(
    private val reloj: Reloj,
) {
    operator fun invoke(reserva: Reserva): Boolean =
        reserva.estado == EstadoReserva.CONFIRMADA &&
            reserva.inicio - reloj.ahora() >= MARGEN_CANCELACION
}

class CancelarReserva(
    private val repositorio: RepositorioReservas,
    private val sePuedeCancelar: SePuedeCancelar,
) {
    suspend operator fun invoke(
        socioId: String,
        reserva: Reserva,
    ): Resultado<Reserva> {
        if (!sePuedeCancelar(reserva)) {
            return Resultado.Rechazo("cancelacion_fuera_de_plazo", "Solo se puede cancelar hasta 2 horas antes del inicio.")
        }
        return repositorio.cancelar(socioId, reserva.id)
    }
}

/** Las reservas confirmadas que aún no han empezado, de la más cercana a la más lejana. */
class ListarMisReservas(
    private val repositorio: RepositorioReservas,
    private val reloj: Reloj,
) {
    suspend operator fun invoke(socioId: String): Resultado<List<Reserva>> =
        when (val resultado = repositorio.reservasDe(socioId)) {
            is Resultado.Exito -> {
                val ahora = reloj.ahora()
                Resultado.Exito(
                    resultado.valor
                        .filter { it.estado == EstadoReserva.CONFIRMADA && it.inicio > ahora }
                        .sortedBy { it.inicio },
                )
            }
            else -> resultado
        }
}
