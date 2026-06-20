"""Resumen de una semana de entrenamientos, refactorizado en funciones pequeñas."""

DISTANCIA_MEDIA_MARATON_KM = 21.1

sesiones = [
    {"fecha": "2026-03-02", "km": 5.2, "minutos": 29, "completada": True},
    {"fecha": "2026-03-04", "km": 8.0, "minutos": 47, "completada": True},
    {"fecha": "2026-03-07", "km": 0, "minutos": 0, "completada": False},
    {"fecha": "2026-03-09", "km": 10.5, "minutos": 58, "completada": True},
]


def sesiones_completadas(sesiones):
    """Devuelve solo las sesiones que se llegaron a hacer."""
    return [sesion for sesion in sesiones if sesion["completada"]]


def ritmo(minutos, km):
    """Minutos por kilómetro."""
    return minutos / km


def mejor_sesion(sesiones):
    """Devuelve la sesión con el ritmo más rápido (menos minutos por km)."""
    return min(sesiones, key=lambda sesion: ritmo(sesion["minutos"], sesion["km"]))


def informe_semanal(sesiones):
    """Construye el texto del resumen. No imprime nada."""
    hechas = sesiones_completadas(sesiones)
    total_km = sum(sesion["km"] for sesion in hechas)
    total_minutos = sum(sesion["minutos"] for sesion in hechas)
    mejor = mejor_sesion(hechas)
    mejor_ritmo = ritmo(mejor["minutos"], mejor["km"])
    lineas = [
        f"Total: {round(total_km, 1)} km",
        f"Ritmo medio: {round(ritmo(total_minutos, total_km), 2)} min/km",
        f"Mejor ritmo: {round(mejor_ritmo, 2)} min/km el {mejor['fecha']}",
    ]
    if total_km >= DISTANCIA_MEDIA_MARATON_KM:
        lineas.append("Esta semana has corrido más que una media maratón")
    return "\n".join(lineas)


if __name__ == "__main__":
    print(informe_semanal(sesiones))
