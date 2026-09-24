package agente

import (
	"github.com/yecharlot/AlsetOS/organismo"
	"github.com/yecharlot/AlsetOS/pulso"
)

// Agente representa un actor autónomo mínimo asociado a un organismo.
type Agente struct {
	Nombre   string
	Objetivo string
}

// Observar lee una señal del organismo.
func (agente Agente) Observar(entidad *organismo.Organismo) string {
	return entidad.Memoria["ultimo_pulso"]
}

// Actuar registra una acción y emite un pulso.
func (agente Agente) Actuar(entidad *organismo.Organismo, accion string) {
	entidad.Memoria["ultima_accion"] = accion
	pulso.Emitir(entidad, pulso.Pulso{
		Tipo:      "agente",
		Origen:    agente.Nombre,
		Contenido: accion,
	})
}
