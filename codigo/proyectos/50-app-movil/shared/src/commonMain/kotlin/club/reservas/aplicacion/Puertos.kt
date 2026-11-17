package club.reservas.aplicacion

import club.reservas.dominio.HoraLibre
import club.reservas.dominio.Pista
import club.reservas.dominio.Reserva
import club.reservas.dominio.Resultado
import kotlinx.datetime.LocalDate
import kotlinx.datetime.TimeZone
import kotlin.time.Clock
import kotlin.time.Instant

/** Lo que la app necesita de la API de reservas. Hay una versión HTTP y otra simulada. */
interface RepositorioReservas {
    suspend fun pistas(): Resultado<List<Pista>>

    suspend fun horasLibres(
        pistaId: String,
        fecha: LocalDate,
    ): Resultado<List<HoraLibre>>

    suspend fun reservar(
        socioId: String,
        pistaId: String,
        inicio: String,
    ): Resultado<Reserva>

    suspend fun cancelar(
        socioId: String,
        reservaId: String,
    ): Resultado<Reserva>

    suspend fun reservasDe(socioId: String): Resultado<List<Reserva>>
}

/** El reloj se inyecta para que los tests puedan fijar "ahora". */
interface Reloj {
    fun ahora(): Instant

    val zona: TimeZone
}

object RelojSistema : Reloj {
    override fun ahora(): Instant = Clock.System.now()

    override val zona: TimeZone = TimeZone.currentSystemDefault()
}
