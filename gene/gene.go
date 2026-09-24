package gene

import (
	"fmt"

	"github.com/yecharlot/AlsetOS/organismo"
)

// Gene representa una capacidad ejecutable.
type Gene struct {
	Nombre string
}

// Ejecutar ejecuta la capacidad y registra el último resultado en memoria.
func (gen Gene) Ejecutar(entidad *organismo.Organismo) string {
	resultado := fmt.Sprintf("Gene %s ejecutado para %s", gen.Nombre, entidad.Nombre)
	entidad.Memoria["ultimo_resultado"] = resultado
	return resultado
}
