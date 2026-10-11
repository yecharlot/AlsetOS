package capacidad

import (
	"testing"

	"github.com/yecharlot/AlsetOS/organismo"
)

func TestOtorgarSoloAutoridad(t *testing.T) {
	maestro := organismo.NuevoFundacional("Maestro", "boot")
	app := organismo.Nuevo("app")
	if err := Otorgar(app, app, "gene.ejecutar"); err == nil {
		t.Fatal("sin autoridad")
	}
	if err := Otorgar(maestro, app, "gene.ejecutar"); err != nil {
		t.Fatal(err)
	}
	if !Tiene(app, "gene.ejecutar") {
		t.Fatal("no otorgada")
	}
}

func TestRequerirDefaultDeny(t *testing.T) {
	o := organismo.Nuevo("x")
	if err := Requerir(o, "wasm.ejecutar"); err == nil {
		t.Fatal("default deny")
	}
}
