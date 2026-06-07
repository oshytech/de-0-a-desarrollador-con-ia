original = [8000, 12000, 6500]
copia = original          # no es una copia: es otro nombre para la misma lista
copia.append(15000)
print(original)

independiente = original.copy()
independiente.append(3000)
print(original)
print(independiente)
