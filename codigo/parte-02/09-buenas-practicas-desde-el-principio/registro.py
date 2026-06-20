"""Logging en lugar de print: niveles, formato y la traza completa de un error."""
import logging

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s: %(message)s",
)
log = logging.getLogger("entrenamientos")


def ritmo(minutos, km):
    log.debug("Calculando ritmo con minutos=%s y km=%s", minutos, km)
    return minutos / km


for minutos, km in [(29, 5.2), (0, 0)]:
    try:
        log.info("Ritmo: %.2f min/km", ritmo(minutos, km))
    except ZeroDivisionError:
        log.exception("Sesión con 0 km: no se puede calcular el ritmo")
