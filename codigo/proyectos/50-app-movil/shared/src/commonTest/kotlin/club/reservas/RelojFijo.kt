package club.reservas

import club.reservas.aplicacion.Reloj
import kotlinx.datetime.TimeZone
import kotlin.time.Instant

/** Lunes 23 de noviembre de 2026 a las 09:30, hora de Madrid. */
class RelojFijo(
    texto: String = "2026-11-23T09:30:00+01:00",
) : Reloj {
    var momento: Instant = Instant.parse(texto)

    override fun ahora(): Instant = momento

    override val zona: TimeZone = TimeZone.of("Europe/Madrid")
}
