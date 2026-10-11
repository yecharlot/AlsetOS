package app

import (
	"testing"

	"github.com/yecharlot/AlsetOS/frontera"
	"github.com/yecharlot/AlsetOS/organismo"
)

func TestInstalarYCorrerMinimoEsfuerzo(t *testing.T) {
	rt := frontera.NuevoRuntime()
	maestro := organismo.NuevoFundacional("Maestro", "boot")
	o, err := Instalar(rt, maestro, Definicion{
		Nombre:  "pedidos-ruta",
		Version: "1",
		Metas:   []string{"gastronomia"},
	}, []string{"gene.ejecutar"})
	if err != nil {
		t.Fatal(err)
	}
	if !o.DefinicionVerificada || o.RootCID == "" {
		t.Fatal("definición debe estar verificada")
	}
	if !o.TieneCapacidad("gene.ejecutar") {
		t.Fatal("cap otorgada por autoridad")
	}
	res := Correr(rt, "pedidos-ruta", func(o *organismo.Organismo) error {
		o.Memoria["pedido"] = "1"
		return nil
	})
	if res.Error != "" || !res.KernelVivo {
		t.Fatalf("%+v", res)
	}
}

func TestSinAutoridadNoOtorga(t *testing.T) {
	rt := frontera.NuevoRuntime()
	user := organismo.Nuevo("intruso")
	_, err := Instalar(rt, user, Definicion{Nombre: "x", Version: "1"}, []string{"gene.ejecutar"})
	if err == nil {
		t.Fatal("usuario sin caps.otorgar no debe instalar con caps")
	}
}
