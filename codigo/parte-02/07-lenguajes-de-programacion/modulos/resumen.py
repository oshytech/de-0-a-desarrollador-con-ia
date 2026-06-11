import dinero

gastos = ["12,50", "3,20", "40"]
total = sum(dinero.a_centimos(g) for g in gastos)
print(f"Total: {dinero.formatear(total)}")
