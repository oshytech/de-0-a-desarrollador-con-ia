contador = 3
while contador > 0:
    print(f"Quedan {contador}")
    contador = contador - 1
print("Despegue")

for numero in range(1, 8):
    if numero % 2 == 0:
        continue
    print(numero, end=" ")
print()

semana = [8250, 11020, 6400, 12500]
for pasos in semana:
    if pasos < 7000:
        print(f"Primer día flojo: {pasos} pasos")
        break
