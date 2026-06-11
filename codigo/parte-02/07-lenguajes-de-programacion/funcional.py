viajes = [150, 75, 150, 150, 75, 300]

# Estilo imperativo: dices cómo, paso a paso
caros = []
for precio in viajes:
    if precio >= 150:
        caros.append(precio)
print(caros)

# Estilo funcional: describes la transformación
print([precio for precio in viajes if precio >= 150])
print(list(map(lambda precio: precio / 100, viajes)))
print(sum(viajes) / 100)
