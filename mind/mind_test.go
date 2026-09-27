package mind

import (
	"testing"

	"github.com/yecharlot/AlsetOS/organismo"
	"github.com/yecharlot/AlsetOS/zyrion"
)

func TestDecidirSiEjecutaGene(t *testing.T) {
	o := organismo.Nuevo("x")
	o.Capacidad["gene.ejecutar"] = true
	m := Mente{}
	if m.Decidir(zyrion.Si, o) != "ejecutar_gene" {
		t.Fatal(m.Decidir(zyrion.Si, o))
	}
}

func TestDecidirInciertoDetiene(t *testing.T) {
	o := organismo.Nuevo("x")
	m := Mente{}
	if m.Decidir(zyrion.Incierto, o) != "detener" {
		t.Fatal("expected detener")
	}
	if m.DecidirConEstrategia(zyrion.Incierto, o, Evaluar) != "evaluar_incierto" {
		t.Fatal("expected evaluar")
	}
}
