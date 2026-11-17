package club.reservas.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import club.reservas.aplicacion.CancelarReserva
import club.reservas.aplicacion.ConsultarHorasLibres
import club.reservas.aplicacion.DiasReservables
import club.reservas.aplicacion.ListarMisReservas
import club.reservas.aplicacion.ListarPistas
import club.reservas.aplicacion.Reloj
import club.reservas.aplicacion.RepositorioReservas
import club.reservas.aplicacion.ReservarPista
import club.reservas.aplicacion.SePuedeCancelar
import club.reservas.dominio.HoraLibre
import club.reservas.dominio.Pista
import club.reservas.dominio.Reserva
import club.reservas.dominio.Resultado
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.datetime.LocalDate

/** Todo lo que la pantalla necesita para pintarse. Es inmutable: cada cambio crea una copia. */
data class EstadoPantalla(
    val cargando: Boolean = false,
    val pistas: List<Pista> = emptyList(),
    val pistaElegida: Pista? = null,
    val dias: List<LocalDate> = emptyList(),
    val diaElegido: LocalDate? = null,
    val horas: List<HoraLibre> = emptyList(),
    val misReservas: List<Reserva> = emptyList(),
    val aviso: String? = null,
    val confirmacion: String? = null,
)

class ReservasViewModel(
    repositorio: RepositorioReservas,
    reloj: Reloj,
    private val socioId: String,
) : ViewModel() {
    private val diasReservables = DiasReservables(reloj)
    private val listarPistas = ListarPistas(repositorio)
    private val consultarHoras = ConsultarHorasLibres(repositorio, diasReservables)
    private val reservarPista = ReservarPista(repositorio)
    private val sePuedeCancelar = SePuedeCancelar(reloj)
    private val cancelarReserva = CancelarReserva(repositorio, sePuedeCancelar)
    private val listarMisReservas = ListarMisReservas(repositorio, reloj)

    private val _estado = MutableStateFlow(EstadoPantalla(dias = diasReservables()))
    val estado: StateFlow<EstadoPantalla> = _estado.asStateFlow()

    init {
        viewModelScope.launch {
            _estado.update { it.copy(cargando = true) }
            when (val r = listarPistas()) {
                is Resultado.Exito -> _estado.update { it.copy(pistas = r.valor) }
                else -> avisar(r)
            }
            _estado.update { it.copy(cargando = false) }
            recargarMisReservas()
        }
    }

    fun elegirPista(pista: Pista) {
        _estado.update { it.copy(pistaElegida = pista, horas = emptyList(), confirmacion = null) }
        recargarHoras()
    }

    fun elegirDia(dia: LocalDate) {
        _estado.update { it.copy(diaElegido = dia, horas = emptyList(), confirmacion = null) }
        recargarHoras()
    }

    fun reservar(hora: HoraLibre) {
        val pista = _estado.value.pistaElegida ?: return
        viewModelScope.launch {
            when (val r = reservarPista(socioId, pista.id, hora)) {
                is Resultado.Exito ->
                    _estado.update {
                        it.copy(confirmacion = "Reserva confirmada: ${pista.nombre}, ${r.valor.fechaLocal} a las ${r.valor.horaLocal}.")
                    }
                else -> avisar(r)
            }
            recargarHoras()
            recargarMisReservas()
        }
    }

    fun puedeCancelar(reserva: Reserva): Boolean = sePuedeCancelar(reserva)

    fun cancelar(reserva: Reserva) {
        viewModelScope.launch {
            when (val r = cancelarReserva(socioId, reserva)) {
                is Resultado.Exito -> _estado.update { it.copy(confirmacion = "Reserva cancelada.") }
                else -> avisar(r)
            }
            recargarMisReservas()
            recargarHoras()
        }
    }

    fun cerrarAviso() = _estado.update { it.copy(aviso = null) }

    /** La confirmación habla de la pestaña donde se hizo la acción: al cambiar de pestaña, se quita. */
    fun cerrarConfirmacion() = _estado.update { it.copy(confirmacion = null) }

    private fun recargarHoras() {
        val pista = _estado.value.pistaElegida ?: return
        val dia = _estado.value.diaElegido ?: return
        viewModelScope.launch {
            when (val r = consultarHoras(pista.id, dia)) {
                is Resultado.Exito -> _estado.update { it.copy(horas = r.valor) }
                else -> avisar(r)
            }
        }
    }

    private suspend fun recargarMisReservas() {
        when (val r = listarMisReservas(socioId)) {
            is Resultado.Exito -> _estado.update { it.copy(misReservas = r.valor) }
            else -> avisar(r)
        }
    }

    private fun avisar(resultado: Resultado<*>) {
        val texto =
            when (resultado) {
                is Resultado.Rechazo -> resultado.mensaje
                is Resultado.SinConexion -> "No hay conexión con el club. Revisa la red e inténtalo de nuevo."
                is Resultado.Exito -> return
            }
        _estado.update { it.copy(aviso = texto, confirmacion = null) }
    }
}
