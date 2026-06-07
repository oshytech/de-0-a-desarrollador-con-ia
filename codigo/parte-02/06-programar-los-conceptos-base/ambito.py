def kilometros(pasos, zancada_m=0.75):
    distancia = pasos * zancada_m / 1000
    return distancia


print(kilometros(10000))
print(kilometros(10000, zancada_m=0.8))
print(distancia)
