// Demo del capítulo 37: migra el esquema y enseña por qué la regla
// "una pista no puede tener dos reservas que se solapen" tiene que
// defenderla la base de datos y no solo el código.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	_ "time/tzdata" // incluye las zonas horarias en el binario

	"github.com/jackc/pgx/v5/pgxpool"

	"reservas/internal/adaptadores/postgres"
	"reservas/internal/dominio"
	"reservas/migraciones"
)

func main() {
	ctx := context.Background()
	url := os.Getenv("RESERVAS_BD")
	if url == "" {
		url = "postgres://club:club@localhost:5432/club?sslmode=disable"
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	aplicadas, err := postgres.Migrar(ctx, pool, migraciones.Ficheros)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Migraciones aplicadas ahora:", aplicadas)

	prepararDatos(ctx, pool)

	madrid, _ := time.LoadLocation("Europe/Madrid")
	manana := time.Now().In(madrid).AddDate(0, 0, 1)
	diezEnPunto := time.Date(manana.Year(), manana.Month(), manana.Day(), 10, 0, 0, 0, madrid)

	fmt.Println("\n1) Comprobar y luego insertar, sin restricción en la tabla:")
	fmt.Printf("   filas para la misma pista y hora: %d\n", reservaIngenua(ctx, pool, diezEnPunto))

	fmt.Println("\n2) Cinco socios a la vez contra la tabla con la restricción:")
	repo := postgres.NuevoRepositorio(pool)
	var wg sync.WaitGroup
	resultados := make([]string, 5)
	for i := range 5 {
		wg.Add(1)
		go func(socio int64) {
			defer wg.Done()
			id, err := repo.Reservar(ctx, socio, 1, diezEnPunto)
			if err != nil {
				resultados[socio-1] = fmt.Sprintf("   socio %d: %v", socio, err)
				return
			}
			resultados[socio-1] = fmt.Sprintf("   socio %d: reserva %d confirmada", socio, id)
		}(int64(i + 1))
	}
	wg.Wait()
	for _, r := range resultados {
		fmt.Println(r)
	}

	fmt.Println("\n3) Un socio que intenta una tercera reserva futura:")
	for _, hora := range []int{12, 13, 14} {
		inicio := time.Date(manana.Year(), manana.Month(), manana.Day(), hora, 0, 0, 0, madrid)
		_, err := repo.Reservar(ctx, 2, 1, inicio)
		fmt.Printf("   %02d:00 -> %s\n", hora, describir(err))
	}

	fmt.Println("\n4) Una reserva a las 07:00:")
	siete := time.Date(manana.Year(), manana.Month(), manana.Day(), 7, 0, 0, 0, madrid)
	_, err = repo.Reservar(ctx, 3, 1, siete)
	fmt.Println("  ", describir(err))
}

func describir(err error) string {
	switch {
	case err == nil:
		return "confirmada"
	case errors.Is(err, dominio.ErrHorarioNoValido):
		return dominio.ErrHorarioNoValido.Error()
	default:
		return err.Error()
	}
}

// prepararDatos deja cinco socios, una pista y ninguna reserva.
func prepararDatos(ctx context.Context, pool *pgxpool.Pool) {
	_, err := pool.Exec(ctx, `
		TRUNCATE reserva;
		INSERT INTO socio (nombre, correo) VALUES
			('Ana', 'ana@ejemplo.org'), ('Luis', 'luis@ejemplo.org'),
			('Marta', 'marta@ejemplo.org'), ('Pablo', 'pablo@ejemplo.org'),
			('Sara', 'sara@ejemplo.org')
		ON CONFLICT (correo) DO NOTHING;
		INSERT INTO pista (nombre, deporte) VALUES ('Pista 1', 'padel')
		ON CONFLICT (nombre) DO NOTHING;
		CREATE TABLE IF NOT EXISTS reserva_ingenua (
			pista_id bigint, inicio timestamptz, fin timestamptz);
		TRUNCATE reserva_ingenua;`)
	if err != nil {
		log.Fatal(err)
	}
}

// reservaIngenua lanza cinco peticiones que comprueban si la pista está
// libre y, si lo está, insertan. Entre comprobar e insertar hay una pausa
// de 50 ms para que la carrera se vea siempre: en producción el hueco es
// más pequeño, pero existe igual.
func reservaIngenua(ctx context.Context, pool *pgxpool.Pool, inicio time.Time) int {
	fin := inicio.Add(time.Hour)
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var ocupada bool
			pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM reserva_ingenua
				WHERE pista_id = 1 AND tstzrange(inicio, fin) && tstzrange($1, $2))`,
				inicio, fin).Scan(&ocupada)
			if ocupada {
				return
			}
			time.Sleep(50 * time.Millisecond)
			pool.Exec(ctx, "INSERT INTO reserva_ingenua VALUES (1, $1, $2)", inicio, fin)
		}()
	}
	wg.Wait()
	var n int
	pool.QueryRow(ctx, "SELECT count(*) FROM reserva_ingenua").Scan(&n)
	return n
}
