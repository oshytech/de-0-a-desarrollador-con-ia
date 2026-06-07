from decimal import Decimal

print(0.1 + 0.2)
print(0.1 + 0.2 == 0.3)

# Dos formas de evitar el problema con dinero
print(10 + 20)  # trabajar en céntimos, con enteros
print(Decimal("0.1") + Decimal("0.2"))
