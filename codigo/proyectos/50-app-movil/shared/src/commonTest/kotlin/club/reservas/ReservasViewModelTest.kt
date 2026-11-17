package club.reservas

import club.reservas.datos.RepositorioSimulado
import club.reservas.ui.ReservasViewModel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import kotlinx.datetime.LocalDate
import kotlin.test.AfterTest
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull

@OptIn(ExperimentalCoroutinesApi::class)
class ReservasViewModelTest {
    @BeforeTest
    fun antes() = Dispatchers.setMain(UnconfinedTestDispatcher())

    @AfterTest
    fun despues() = Dispatchers.resetMain()

    @Test
    fun reservarUnaHoraLaQuitaDeLasLibresYLaAnadeAMisReservas() =
        runTest {
            val vm = ReservasViewModel(RepositorioSimulado(RelojFijo()), RelojFijo(), socioId = "s-1")
            vm.elegirPista(
                vm.estado.value.pistas
                    .first(),
            )
            vm.elegirDia(LocalDate(2026, 11, 24))
            val primera =
                vm.estado.value.horas
                    .first()

            vm.reservar(primera)

            val estado = vm.estado.value
            assertNull(estado.aviso)
            assertEquals("Reserva confirmada: Pista 1, 2026-11-24 a las 08:00.", estado.confirmacion)
            assertEquals("09:00", estado.horas.first().etiqueta)
            assertEquals(1, estado.misReservas.size)
        }

    @Test
    fun unRechazoDeLaApiSeEnsenaComoAviso() =
        runTest {
            val vm = ReservasViewModel(RepositorioSimulado(RelojFijo()), RelojFijo(), socioId = "s-1")
            vm.elegirPista(
                vm.estado.value.pistas
                    .first(),
            )
            vm.elegirDia(LocalDate(2026, 11, 24))
            val horas = vm.estado.value.horas
            vm.reservar(horas[0])
            vm.reservar(horas[1])
            vm.reservar(horas[2])

            assertEquals("Ya tienes 2 reservas pendientes. Cancela una para reservar otra.", vm.estado.value.aviso)
        }

    @Test
    fun cerrarLaConfirmacionNoTocaLasReservas() =
        runTest {
            val vm = ReservasViewModel(RepositorioSimulado(RelojFijo()), RelojFijo(), socioId = "s-1")
            vm.elegirPista(
                vm.estado.value.pistas
                    .first(),
            )
            vm.elegirDia(LocalDate(2026, 11, 24))
            vm.reservar(
                vm.estado.value.horas
                    .first(),
            )

            vm.cerrarConfirmacion()

            assertNull(vm.estado.value.confirmacion)
            assertEquals(1, vm.estado.value.misReservas.size)
        }
}
