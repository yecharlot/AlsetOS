package pulso

import (
	"fmt"
	"time"
	"github.com/yecharlot/AlsetOS/organismo"
)

type Pulso struct {
	Tipo string
	Origen string
	Contenido string
	Fecha time.Time
}

func Emitir(entidad *organismo.Organismo, evento Pulso) {
	entidad.Memoria["ultimo_pulso"] = evento.Tipo + ":" + evento.Contenido
	fmt.Printf("[PULSO] %s -> %s | %s\n", evento.Origen, evento.Tipo, evento.Contenido)
}

// Recibir entrega un pulso a un canal sin crear dependencia de red.
func Recibir(canal chan Pulso, evento Pulso) {
	canal <- evento
}
