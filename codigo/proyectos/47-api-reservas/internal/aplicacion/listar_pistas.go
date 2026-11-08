package aplicacion

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// PistaDTO es una pista tal como sale hacia fuera.
type PistaDTO struct {
	ID      string
	Nombre  string
	Deporte string
}

// ListarPistas devuelve las pistas del club ordenadas por id.
type ListarPistas struct {
	pistas RepositorioPistas
}

func NuevoListarPistas(pistas RepositorioPistas) *ListarPistas {
	return &ListarPistas{pistas: pistas}
}

func (uc *ListarPistas) Ejecutar(ctx context.Context) ([]PistaDTO, error) {
	todas, err := uc.pistas.Todas(ctx)
	if err != nil {
		return nil, fmt.Errorf("pistas: %w", err)
	}
	resultado := make([]PistaDTO, 0, len(todas))
	for _, p := range todas {
		resultado = append(resultado, PistaDTO{ID: p.ID, Nombre: p.Nombre, Deporte: string(p.Deporte)})
	}
	slices.SortFunc(resultado, func(a, b PistaDTO) int { return strings.Compare(a.ID, b.ID) })
	return resultado, nil
}
