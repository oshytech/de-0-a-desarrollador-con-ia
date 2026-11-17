package club.reservas.datos

import club.reservas.aplicacion.RepositorioReservas
import club.reservas.dominio.HoraLibre
import club.reservas.dominio.Pista
import club.reservas.dominio.Reserva
import club.reservas.dominio.Resultado
import io.ktor.client.HttpClient
import io.ktor.client.HttpClientConfig
import io.ktor.client.call.body
import io.ktor.client.engine.HttpClientEngine
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.get
import io.ktor.client.request.header
import io.ktor.client.request.parameter
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.http.isSuccess
import io.ktor.serialization.kotlinx.json.json
import kotlinx.coroutines.CancellationException
import kotlinx.datetime.LocalDate
import kotlinx.serialization.json.Json

private const val CABECERA_SOCIO = "X-Socio-Id"

/**
 * Cliente de la API de reservas con Ktor. El mismo código se usa en Android y en iOS:
 * cada plataforma aporta su motor HTTP (OkHttp en Android, Darwin en iOS).
 */
class RepositorioHttp(
    private val urlBase: String,
    motor: HttpClientEngine? = null,
) : RepositorioReservas {
    private val cliente: HttpClient =
        if (motor == null) HttpClient { configurar() } else HttpClient(motor) { configurar() }

    private fun HttpClientConfig<*>.configurar() {
        expectSuccess = false
        install(ContentNegotiation) { json(Json { ignoreUnknownKeys = true }) }
        install(HttpTimeout) { requestTimeoutMillis = 10_000 }
    }

    override suspend fun pistas(): Resultado<List<Pista>> =
        llamar({ get("$urlBase/pistas") }) { it.body<List<PistaDto>>().map(PistaDto::aDominio) }

    override suspend fun horasLibres(
        pistaId: String,
        fecha: LocalDate,
    ): Resultado<List<HoraLibre>> =
        llamar({
            get("$urlBase/pistas/$pistaId/disponibilidad") { parameter("fecha", fecha.toString()) }
        }) { it.body<DisponibilidadDto>().aDominio() }

    override suspend fun reservar(
        socioId: String,
        pistaId: String,
        inicio: String,
    ): Resultado<Reserva> =
        llamar({
            post("$urlBase/reservas") {
                header(CABECERA_SOCIO, socioId)
                contentType(ContentType.Application.Json)
                setBody(NuevaReservaDto(pistaId, inicio))
            }
        }) { it.body<ReservaDto>().aDominio() }

    override suspend fun cancelar(
        socioId: String,
        reservaId: String,
    ): Resultado<Reserva> =
        llamar({
            post("$urlBase/reservas/$reservaId/cancelacion") { header(CABECERA_SOCIO, socioId) }
        }) { it.body<ReservaDto>().aDominio() }

    override suspend fun reservasDe(socioId: String): Resultado<List<Reserva>> =
        llamar({ get("$urlBase/socios/$socioId/reservas") { header(CABECERA_SOCIO, socioId) } }) {
            it.body<List<ReservaDto>>().map(ReservaDto::aDominio)
        }

    /** Traduce cualquier respuesta o fallo de red a un [Resultado]. La interfaz nunca ve excepciones. */
    private suspend fun <T> llamar(
        peticion: suspend HttpClient.() -> HttpResponse,
        convertir: suspend (HttpResponse) -> T,
    ): Resultado<T> {
        val respuesta =
            try {
                cliente.peticion()
            } catch (e: CancellationException) {
                throw e
            } catch (e: Exception) {
                return Resultado.SinConexion(e.message ?: "Sin conexión con el club")
            }
        if (respuesta.status.isSuccess()) return Resultado.Exito(convertir(respuesta))

        val error = runCatching { respuesta.body<ErrorDto>() }.getOrNull()
        return Resultado.Rechazo(
            codigo = error?.error ?: "http_${respuesta.status.value}",
            mensaje = error?.mensaje ?: mensajePorDefecto(respuesta.status.value),
        )
    }

    private fun mensajePorDefecto(estado: Int) =
        when (estado) {
            404 -> "No se ha encontrado lo que buscabas."
            409 -> "No se ha podido completar: la pista o la reserva han cambiado."
            422 -> "La reserva no cumple las normas del club."
            else -> "El servidor ha respondido con un error ($estado)."
        }
}
