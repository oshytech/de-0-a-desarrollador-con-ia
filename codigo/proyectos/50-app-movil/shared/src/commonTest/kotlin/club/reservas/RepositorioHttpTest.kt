package club.reservas

import club.reservas.datos.RepositorioHttp
import club.reservas.dominio.EstadoReserva
import club.reservas.dominio.Resultado
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.client.engine.mock.toByteArray
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpMethod
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import kotlinx.coroutines.test.runTest
import kotlinx.datetime.LocalDate
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs

class RepositorioHttpTest {
    private val json = headersOf(HttpHeaders.ContentType, "application/json")
    private val reservaJson =
        """{"id":"r-7","socio_id":"s-1","pista_id":"p-1","inicio":"2026-11-24T10:00:00+01:00",""" +
            """"fin":"2026-11-24T11:00:00+01:00","estado":"confirmada"}"""

    @Test
    fun leeLasPistas() =
        runTest {
            val motor =
                MockEngine { peticion ->
                    assertEquals("/pistas", peticion.url.encodedPath)
                    respond("""[{"id":"p-1","nombre":"Pista 1","deporte":"tenis"}]""", HttpStatusCode.OK, json)
                }
            val resultado = RepositorioHttp("http://api.local", motor).pistas()
            assertEquals("Pista 1", (resultado as Resultado.Exito).valor.single().nombre)
        }

    @Test
    fun convierteLaDisponibilidadEnHorasListasParaReservar() =
        runTest {
            val motor =
                MockEngine { peticion ->
                    assertEquals("2026-11-24", peticion.url.parameters["fecha"])
                    respond("""{"fecha":"2026-11-24","horas_libres":["08:00","09:00"]}""", HttpStatusCode.OK, json)
                }
            val resultado = RepositorioHttp("http://api.local", motor).horasLibres("p-1", LocalDate(2026, 11, 24))
            assertEquals(listOf("2026-11-24T08:00", "2026-11-24T09:00"), (resultado as Resultado.Exito).valor.map { it.inicio })
        }

    @Test
    fun reservaComoElSocioDeLaCabecera() =
        runTest {
            val motor =
                MockEngine { peticion ->
                    assertEquals("s-1", peticion.headers["X-Socio-Id"])
                    assertEquals("""{"pista_id":"p-1","inicio":"2026-11-24T10:00"}""", peticion.body.toByteArray().decodeToString())
                    respond(reservaJson, HttpStatusCode.Created, json)
                }
            val resultado = RepositorioHttp("http://api.local", motor).reservar("s-1", "p-1", "2026-11-24T10:00")
            assertEquals("10:00", (resultado as Resultado.Exito).valor.horaLocal)
        }

    @Test
    fun unConflictoLlegaComoRechazoConElMensajeDeLaApi() =
        runTest {
            val motor =
                MockEngine {
                    respond(
                        """{"error":"pista_ocupada","mensaje":"La pista ya está reservada a esa hora."}""",
                        HttpStatusCode.Conflict,
                        json,
                    )
                }
            val resultado = RepositorioHttp("http://api.local", motor).reservar("s-1", "p-1", "2026-11-24T10:00")
            assertIs<Resultado.Rechazo>(resultado)
            assertEquals("pista_ocupada", resultado.codigo)
            assertEquals("La pista ya está reservada a esa hora.", resultado.mensaje)
        }

    @Test
    fun cancelarUsaLaRutaDeCancelacion() =
        runTest {
            val motor =
                MockEngine { peticion ->
                    assertEquals(HttpMethod.Post, peticion.method)
                    assertEquals("/reservas/r-7/cancelacion", peticion.url.encodedPath)
                    assertEquals("s-1", peticion.headers["X-Socio-Id"])
                    respond(reservaJson.replace("confirmada", "cancelada"), HttpStatusCode.OK, json)
                }
            val resultado = RepositorioHttp("http://api.local", motor).cancelar("s-1", "r-7")
            assertEquals(EstadoReserva.CANCELADA, (resultado as Resultado.Exito).valor.estado)
        }

    @Test
    fun sinRedDevuelveSinConexionEnLugarDeRomper() =
        runTest {
            val motor = MockEngine { error("Conexión rechazada") }
            assertIs<Resultado.SinConexion>(RepositorioHttp("http://api.local", motor).pistas())
        }
}
