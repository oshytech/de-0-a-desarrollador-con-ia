package dominio

type Deporte string

const (
	Tenis Deporte = "tenis"
	Padel Deporte = "padel"
)

type Pista struct {
	ID      string
	Nombre  string
	Deporte Deporte
}
