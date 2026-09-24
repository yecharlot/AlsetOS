package mind

import (
	"github.com/yecharlot/AlsetOS/organismo"
	"github.com/yecharlot/AlsetOS/zyrion"
)

// Mente coordina observación, decisión y acción.
type Mente struct{}

// Decidir determina si el Gene puede ejecutarse según Zyrion y capacidades.
func (mente Mente) Decidir(estado zyrion.Valor, entidad *organismo.Organismo) string {
	if estado == zyrion.Si && entidad.Capacidad["gene.ejecutar"] {
		return "ejecutar_gene"
	}
	return "detener"
}
