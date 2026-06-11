"""Tres esperas simuladas, primero una detrás de otra y después a la vez."""
import asyncio
import time


async def descargar(nombre):
    await asyncio.sleep(1)  # simula esperar a la red durante un segundo
    return f"{nombre} descargado"


async def una_detras_de_otra():
    return [await descargar(n) for n in ("a", "b", "c")]


async def a_la_vez():
    return await asyncio.gather(descargar("a"), descargar("b"), descargar("c"))


for estrategia in (una_detras_de_otra, a_la_vez):
    inicio = time.perf_counter()
    resultados = asyncio.run(estrategia())
    print(f"{estrategia.__name__}: {time.perf_counter() - inicio:.1f} s {resultados}")
