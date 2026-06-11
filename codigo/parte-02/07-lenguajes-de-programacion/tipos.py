def total_con_iva(precio):
    return precio * 1.21


print(total_con_iva(100))
print(total_con_iva("100"))  # el error solo aparece al llegar aquí
