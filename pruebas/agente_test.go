package pruebas

import (
	"testing"
	"github.com/yecharlot/AlsetOS/agente"
	"github.com/yecharlot/AlsetOS/organismo"
)

func TestAgenteActua(t *testing.T) {
	entidad := organismo.Nuevo("prueba")
	actor := agente.Agente{Nombre:"supervisor"}
	actor.Actuar(entidad, "observar")
	if entidad.Memoria["ultima_accion"] != "observar" { t.Fatal("acción no registrada") }
}
