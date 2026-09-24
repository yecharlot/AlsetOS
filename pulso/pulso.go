package pulso

import (
	"fmt"
	"time"

	"github.com/yecharlot/AlsetOS/organismo"
)

// Pulso representa comunicación y trazabilidad mínima.
type Pulso struct {
	Tipo      string
	Origen    string
	Contenido string
	Fecha     time.Time
}

// Emitir registra el pulso en memoria y lo muestra en la salida.
func Emitir(entidad *organismo.Organismo, evento Pulso) {
	entidad.Memoria["ultimo_pulso"] = evento.Tipo + ":" + evento.Contenido
	fmt.Printf("[PULSO] %s -> %s | %s\n", evento.Origen, evento.Tipo, evento.Contenido)
}
