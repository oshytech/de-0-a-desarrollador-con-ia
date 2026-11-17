package club.reservas.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.PrimaryTabRow
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Tab
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.semantics.LiveRegionMode
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.liveRegion
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import club.reservas.aplicacion.Reloj
import club.reservas.aplicacion.RepositorioReservas
import club.reservas.dominio.Pista
import club.reservas.dominio.Reserva
import kotlinx.datetime.LocalDate

private val DIAS_SEMANA = listOf("lun", "mar", "mié", "jue", "vie", "sáb", "dom")

fun etiquetaDia(fecha: LocalDate): String {
    val texto = fecha.toString()
    return "${DIAS_SEMANA[fecha.dayOfWeek.ordinal]} ${texto.substring(8, 10)}/${texto.substring(5, 7)}"
}

/** Punto de entrada compartido. Android e iOS solo tienen que llamar a esta función. */
@Composable
fun App(
    repositorio: RepositorioReservas,
    reloj: Reloj,
    socioId: String,
) {
    val vm = viewModel { ReservasViewModel(repositorio, reloj, socioId) }
    val estado by vm.estado.collectAsStateWithLifecycle()
    val avisos = remember { SnackbarHostState() }
    var pestana by remember { mutableIntStateOf(0) }
    val cambiarA = { nueva: Int ->
        if (nueva != pestana) vm.cerrarConfirmacion()
        pestana = nueva
    }

    LaunchedEffect(estado.aviso) {
        estado.aviso?.let {
            avisos.showSnackbar(it)
            vm.cerrarAviso()
        }
    }

    MaterialTheme {
        Scaffold(snackbarHost = { SnackbarHost(avisos) }) { margen ->
            Column(Modifier.fillMaxSize().padding(margen)) {
                PrimaryTabRow(selectedTabIndex = pestana) {
                    Tab(selected = pestana == 0, onClick = { cambiarA(0) }, text = { Text("Reservar") })
                    Tab(
                        selected = pestana == 1,
                        onClick = { cambiarA(1) },
                        text = { Text("Mis reservas (${estado.misReservas.size})") },
                    )
                }
                estado.confirmacion?.let {
                    Text(
                        it,
                        modifier =
                            Modifier
                                .padding(16.dp)
                                .semantics { liveRegion = LiveRegionMode.Polite },
                        color = MaterialTheme.colorScheme.primary,
                    )
                }
                if (estado.cargando) {
                    CircularProgressIndicator(Modifier.padding(32.dp).align(Alignment.CenterHorizontally))
                } else if (pestana == 0) {
                    PantallaReservar(estado, vm)
                } else {
                    PantallaMisReservas(estado.misReservas, estado.pistas, vm)
                }
            }
        }
    }
}

@Composable
private fun Titulo(texto: String) {
    Text(
        texto,
        style = MaterialTheme.typography.titleMedium,
        modifier = Modifier.padding(top = 16.dp, bottom = 8.dp).semantics { heading() },
    )
}

@Composable
private fun PantallaReservar(
    estado: EstadoPantalla,
    vm: ReservasViewModel,
) {
    LazyColumn(contentPadding = PaddingValues(horizontal = 16.dp)) {
        item { Titulo("Pista") }
        item {
            LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                items(estado.pistas) { pista ->
                    FilterChip(
                        selected = pista == estado.pistaElegida,
                        onClick = { vm.elegirPista(pista) },
                        label = { Text("${pista.nombre} (${pista.deporte})") },
                    )
                }
            }
        }
        item { Titulo("Día") }
        item {
            LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                items(estado.dias) { dia ->
                    FilterChip(
                        selected = dia == estado.diaElegido,
                        onClick = { vm.elegirDia(dia) },
                        label = { Text(etiquetaDia(dia)) },
                    )
                }
            }
        }
        item { Titulo("Horas libres") }
        when {
            estado.pistaElegida == null || estado.diaElegido == null ->
                item { Text("Elige una pista y un día para ver las horas libres.") }
            estado.horas.isEmpty() ->
                item { Text("No quedan horas libres ese día en esta pista.") }
            else ->
                items(estado.horas) { hora ->
                    Button(onClick = { vm.reservar(hora) }, modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp)) {
                        Text("Reservar a las ${hora.etiqueta}")
                    }
                }
        }
        item { Spacer(Modifier.height(24.dp)) }
    }
}

@Composable
private fun PantallaMisReservas(
    reservas: List<Reserva>,
    pistas: List<Pista>,
    vm: ReservasViewModel,
) {
    if (reservas.isEmpty()) {
        Text("No tienes reservas pendientes.", modifier = Modifier.padding(16.dp))
        return
    }
    LazyColumn(contentPadding = PaddingValues(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        items(reservas, key = { it.id }) { reserva ->
            Card(Modifier.fillMaxWidth()) {
                Row(Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(Modifier.weight(1f)) {
                        Text(
                            pistas.find { it.id == reserva.pistaId }?.nombre ?: reserva.pistaId,
                            style = MaterialTheme.typography.titleMedium,
                        )
                        Text("${reserva.fechaLocal} a las ${reserva.horaLocal}")
                    }
                    val sePuede = vm.puedeCancelar(reserva)
                    OutlinedButton(onClick = { vm.cancelar(reserva) }, enabled = sePuede) {
                        Text(if (sePuede) "Cancelar" else "Ya no se puede cancelar")
                    }
                }
            }
        }
    }
}
