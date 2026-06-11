def total_con_iva(precio: float) -> float:
    return precio * 1.21


print(total_con_iva(100))
print(total_con_iva("100"))  # un comprobador de tipos lo marca sin ejecutar nada
