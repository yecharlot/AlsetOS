package zyrion

// Valor representa la lógica ternaria nativa de AlsetOS.
// 0 = no, 1 = sí, 2 = incierto.
type Valor int

const (
	No Valor = iota
	Si
	Incierto
)

// Texto devuelve la representación humana del valor ternario.
func (valor Valor) Texto() string {
	switch valor {
	case No:
		return "no"
	case Si:
		return "si"
	default:
		return "incierto"
	}
}
