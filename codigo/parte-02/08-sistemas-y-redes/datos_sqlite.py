"""Una base de datos SQLite en memoria: tablas, filas y una consulta."""
import sqlite3

conexion = sqlite3.connect(":memory:")
conexion.execute("CREATE TABLE mediciones (ciudad TEXT, dia TEXT, temperatura REAL)")
conexion.executemany(
    "INSERT INTO mediciones VALUES (?, ?, ?)",
    [
        ("Zaragoza", "lunes", 14.5),
        ("Zaragoza", "martes", 16.0),
        ("Vigo", "lunes", 12.0),
        ("Vigo", "martes", 13.5),
    ],
)
consulta = "SELECT ciudad, AVG(temperatura) FROM mediciones GROUP BY ciudad ORDER BY ciudad"
for ciudad, media in conexion.execute(consulta):
    print(f"{ciudad}: {media:.2f}")
