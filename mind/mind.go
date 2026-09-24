package mind

import (
	"github.com/yecharlot/AlsetOS/organismo"
	"github.com/yecharlot/AlsetOS/zyrion"
)

// EstrategiaIncierto define qué hace la Mente cuando Zyrion no puede determinar
// una condición como sí o no.
type EstrategiaIncierto string

const (
	Detener EstrategiaIncierto = "detener"
	Evaluar  EstrategiaIncierto = "evaluar"
)

// Mente coordina observación, decisión y acción.
type Mente struct{}

// Decidir conserva el comportamiento seguro por defecto:
// un estado incierto se detiene hasta que exista una estrategia explícita.
func (mente Mente) Decidir(estado zyrion.Valor, entidad *organismo.Organismo) string {
	return mente.DecidirConEstrategia(estado, entidad, Detener)
}

// DecidirConEstrategia incorpora incertidumbre como estado de decisión,
// no como un alias implícito de "no".
func (mente Mente) DecidirConEstrategia(estado zyrion.Valor, entidad *organismo.Organismo, estrategia EstrategiaIncierto) string {
	switch estado {
	case zyrion.Si:
		if entidad.Capacidad["gene.ejecutar"] {
			return "ejecutar_gene"
		}
		return "detener"
	case zyrion.Incierto:
		if estrategia == Evaluar {
			return "evaluar_incierto"
		}
		return "detener"
	default:
		return "detener"
	}
}
