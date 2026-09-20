def informe(reservas, pistas):
    # Ocupación de cada pista en un día.
    lineas = []
    for p in pistas:
        n = 0
        for r in reservas:
            if r["pista"] == p["id"] and r["estado"] == "confirmada":
                n = n + 1
        pct = n * 100 / 14
        if pct > 75:
            lineas.append(p["nombre"] + ": " + str(round(pct)) + "% (llena)")
        else:
            lineas.append(p["nombre"] + ": " + str(round(pct)) + "%")
    return "\n".join(lineas)
