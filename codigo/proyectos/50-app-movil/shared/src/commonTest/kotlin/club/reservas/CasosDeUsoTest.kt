package club.reservas

import club.reservas.aplicacion.CancelarReserva
import club.reservas.aplicacion.ConsultarHorasLibres
import club.reservas.aplicacion.DiasReservables
import club.reservas.aplicacion.ListarMisReservas
import club.reservas.aplicacion.SePuedeCancelar
import club.reservas.datos.RepositorioSimulado
import club.reservas.dominio.EstadoReserva
import club.reservas.dominio.Reserva
import club.reservas.dominio.Resultado
import kotlinx.coroutines.test.runTest
import kotlinx.datetime.LocalDate
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertIs
import kotlin.test.assertTrue
import kotlin.time.Duration.Companion.hours
import kotlin.time.Instant

class CasosDeUsoTest {
    private val reloj = RelojFijo()

    private fun reserva(
        inicio: String,
        estado: EstadoReserva = EstadoReserva.CONFIRMADA,
    ) = Reserva(
        id = "r-99",
        socioId = "s-1",
        pistaId = "p-1",
        inicio = Instant.parse(inicio),
        fin = Instant.parse(inicio) + 1.hours,
        estado = estado,
        fechaLocal = inicio.substring(0, 10),
        horaLocal = inicio.substring(11, 16),
    )

    private fun codigo(resultado: Resultado<*>) = (resultado as Resultado.Rechazo).codigo

    @Test
    fun ofreceDesdeHoyHastaDentroDeCatorceDias() {
        val dias = DiasReservables(reloj)()
        assertEquals(15, dias.size)
        assertEquals(LocalDate(2026, 11, 23), dias.first())
        assertEquals(LocalDate(2026, 12, 7), dias.last())
    }

    @Test
    fun noConsultaUnaFechaFueraDePlazo() =
        runTest {
            val consultar = ConsultarHorasLibres(RepositorioSimulado(reloj), DiasReservables(reloj))
            assertEquals("fuera_de_plazo", codigo(consultar("p-1", LocalDate(2026, 12, 8))))
        }

    @Test
    fun noOfreceHorasQueYaHanPasado() =
        runTest {
            val consultar = ConsultarHorasLibres(RepositorioSimulado(reloj), DiasReservables(reloj))
            val horas = (consultar("p-1", LocalDate(2026, 11, 23)) as Resultado.Exito).valor
            assertEquals("2026-11-23T10:00", horas.first().inicio)
            assertEquals("21:00", horas.last().etiqueta)
        }

    @Test
    fun seCancelaConDosHorasDeMargenPeroNoConMenos() {
        val sePuedeCancelar = SePuedeCancelar(reloj)
        assertTrue(sePuedeCancelar(reserva("2026-11-23T11:30:00+01:00")))
        assertFalse(sePuedeCancelar(reserva("2026-11-23T11:00:00+01:00")))
        assertFalse(sePuedeCancelar(reserva("2026-11-24T10:00:00+01:00", EstadoReserva.CANCELADA)))
    }

    @Test
    fun cancelarFueraDePlazoNoLlegaALaApi() =
        runTest {
            val cancelar = CancelarReserva(RepositorioSimulado(reloj), SePuedeCancelar(reloj))
            assertEquals("cancelacion_fuera_de_plazo", codigo(cancelar("s-1", reserva("2026-11-23T10:00:00+01:00"))))
        }

    @Test
    fun misReservasSonLasPendientesOrdenadas() =
        runTest {
            val repositorio = RepositorioSimulado(reloj)
            repositorio.reservar("s-1", "p-2", "2026-11-25T18:00")
            repositorio.reservar("s-1", "p-1", "2026-11-24T10:00")
            val lista = (ListarMisReservas(repositorio, reloj)("s-1") as Resultado.Exito).valor
            assertEquals(listOf("2026-11-24", "2026-11-25"), lista.map { it.fechaLocal })
        }

    @Test
    fun laApiSimuladaAplicaLasReglasDelClub() =
        runTest {
            val api = RepositorioSimulado(reloj)
            assertIs<Resultado.Exito<Reserva>>(api.reservar("s-1", "p-1", "2026-11-24T10:00"))
            assertEquals("pista_ocupada", codigo(api.reservar("s-3", "p-1", "2026-11-24T10:00")))
            assertEquals("horario_no_valido", codigo(api.reservar("s-3", "p-1", "2026-11-24T10:30")))
            assertEquals("fuera_de_plazo", codigo(api.reservar("s-3", "p-1", "2026-12-08T10:00")))
            assertEquals("socio_inactivo", codigo(api.reservar("s-2", "p-2", "2026-11-24T10:00")))
            api.reservar("s-1", "p-2", "2026-11-24T11:00")
            assertEquals("limite_de_reservas", codigo(api.reservar("s-1", "p-2", "2026-11-24T12:00")))
        }

    @Test
    fun laReservaDeOtroSocioSeTrataComoSiNoExistiera() =
        runTest {
            val api = RepositorioSimulado(reloj)
            val reserva = (api.reservar("s-1", "p-1", "2026-11-24T10:00") as Resultado.Exito).valor
            assertEquals("no_encontrado", codigo(api.cancelar("s-3", reserva.id)))
            assertIs<Resultado.Exito<Reserva>>(api.cancelar("s-1", reserva.id))
            assertEquals("reserva_ya_cancelada", codigo(api.cancelar("s-1", reserva.id)))
        }
}
