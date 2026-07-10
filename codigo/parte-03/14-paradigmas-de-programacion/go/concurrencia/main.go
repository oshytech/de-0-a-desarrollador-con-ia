// Concurrencia en Go: enviar recordatorios uno detrás de otro o todos a la vez.
package main

import (
	"fmt"
	"sync"
	"time"
)

// enviarRecordatorio simula una llamada lenta a un servicio de correo.
func enviarRecordatorio(socio string) {
	time.Sleep(200 * time.Millisecond)
}

func main() {
	socios := []string{"Lucía", "Marcos", "Inés", "Jorge"}

	inicio := time.Now()
	for _, s := range socios {
		enviarRecordatorio(s)
	}
	fmt.Println("En secuencia:", time.Since(inicio).Round(100*time.Millisecond))

	inicio = time.Now()
	var wg sync.WaitGroup
	for _, s := range socios {
		wg.Go(func() { enviarRecordatorio(s) })
	}
	wg.Wait()
	fmt.Println("Concurrente:", time.Since(inicio).Round(100*time.Millisecond))
}
