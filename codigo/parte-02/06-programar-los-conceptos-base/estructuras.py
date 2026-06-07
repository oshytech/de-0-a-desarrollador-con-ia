semana = [8250, 11020, 6400]          # lista: ordenada, se puede cambiar
semana.append(12500)
print(semana[0], semana[-1], len(semana))

zancada = (0.75, "metros")            # tupla: ordenada, no se puede cambiar

perfil = {"nombre": "Lucía", "objetivo": 10000}   # diccionario: clave -> valor
perfil["objetivo"] = 12000
print(perfil["nombre"], perfil["objetivo"])

rutas = {"parque", "río", "parque", "centro"}     # conjunto: sin repetidos
print(len(rutas), "parque" in rutas)
