package pruebas

import (
	"testing"
	"github.com/yecharlot/AlsetOS/mind"
	"github.com/yecharlot/AlsetOS/organismo"
	"github.com/yecharlot/AlsetOS/zyrion"
)

func TestMindInciertoConEstrategia(t *testing.T) {
	entidad := organismo.Nuevo("prueba")
	mente := mind.Mente{}
	if decision := mente.DecidirConEstrategia(zyrion.Incierto, entidad, mind.Evaluar); decision != "evaluar_incierto" {
		t.Fatalf("decisión inesperada: %s", decision)
	}
}
