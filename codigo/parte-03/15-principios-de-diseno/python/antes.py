"""Antes: una función que lo hace todo. Funciona, pero tiene muchas razones para cambiar."""
from datetime import datetime, timedelta


def reservar(socio, pista, inicio, reservas, ahora):
    if not socio["activo"]:
        return "Error: socio inactivo"
    if inicio.minute != 0 or inicio.hour < 8 or inicio.hour > 21:
        return "Error: horario no válido"
    if inicio <= ahora or inicio - ahora > timedelta(days=14):
        return "Error: fecha fuera de plazo"
    futuras = 0
    for r in reservas:
        # todas las reservas duran una hora y empiezan en punto:
        # dos reservas se solapan si empiezan a la misma hora
        if r["pista"] == pista and r["estado"] == "confirmada" and r["inicio"] == inicio:
            return "Error: pista ocupada"
        if r["socio"] == socio["id"] and r["estado"] == "confirmada" and r["inicio"] > ahora:
            futuras += 1
    if futuras >= 2:
        return "Error: máximo 2 reservas"
    reservas.append({"socio": socio["id"], "pista": pista, "inicio": inicio, "estado": "confirmada"})
    print(f"Enviando correo a {socio['correo']}: reserva confirmada el {inicio:%d/%m a las %H:%M}")
    return "OK"


if __name__ == "__main__":
    import sys
    sys.stdout.reconfigure(encoding="utf-8")
    lucia = {"id": 1, "correo": "lucia@example.com", "activo": True}
    reservas = []
    ahora = datetime(2026, 3, 10, 12, 0)
    print(reservar(lucia, "Pádel 1", datetime(2026, 3, 11, 19, 0), reservas, ahora))
    print(reservar(lucia, "Pádel 1", datetime(2026, 3, 11, 19, 0), reservas, ahora))
