# Cuentas

Reparto de gastos compartidos desde la terminal: un viaje, un piso o una cena de grupo.
Registra quién paga cada gasto y entre quiénes se reparte, calcula cuánto debe o le
deben a cada persona y propone los pagos para quedar en paz.

Es el proyecto del capítulo 10 del libro. Solo usa la librería estándar de Python.

## Uso

```bash
python main.py anadir "Casa rural" 360 --paga Ana --entre Ana Bruno Carla
python main.py gastos
python main.py saldos
python main.py liquidar
```

Los importes se escriben en euros, con coma o punto y como mucho dos decimales.
Internamente se guardan en céntimos, con números enteros, para que no se pierda
ningún céntimo al sumar. Si un importe no se puede dividir a partes iguales,
los céntimos que sobran se asignan a las primeras personas de la lista.

Los gastos se guardan en `cuentas.json`, en la carpeta desde la que se ejecuta.
Con `--fichero otra_ruta.json` se usa otro fichero, por ejemplo uno por viaje.

## Pruebas

```bash
python -m unittest -v
```

## Estructura

- `reparto.py`: la lógica (importes, reparto, saldos y pagos). Funciones puras.
- `almacen.py`: carga y guarda los gastos en JSON.
- `main.py`: la interfaz de línea de comandos.
- `test_reparto.py`: las pruebas.

## Limitaciones conocidas

- No se pueden borrar ni editar gastos desde el programa.
- Los nombres distinguen mayúsculas: "ana" y "Ana" son dos personas.
- El céntimo sobrante siempre va a las primeras personas de la lista.
