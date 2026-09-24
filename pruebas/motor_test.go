package pruebas

import (
	"testing"

	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/motor"
)

func TestMotorEjecutaOrganismoDeclarativo(t *testing.T) {
	documento := manifiesto.Manifiesto{
		Nombre: "organismo-prueba",
		LispAI: []string{"(recordar origen declarativo)"},
		Zyrion: manifiesto.ConfiguracionZyrion{Estado: "si"},
		Capacidades: []string{"gene.ejecutar"},
		Genes: []string{"gene-saludo"},
	}
	resultado, err := motor.Nuevo().EjecutarManifiesto(documento)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.Decision != "ejecutar_gene" {
		t.Fatalf("decisión inesperada: %s", resultado.Decision)
	}
	if len(resultado.GenesEjecutados) != 1 {
		t.Fatalf("Genes ejecutados: %d", len(resultado.GenesEjecutados))
	}
}

func TestMotorConservaIncertidumbre(t *testing.T) {
	documento := manifiesto.Manifiesto{
		Nombre: "organismo-incierto",
		Zyrion: manifiesto.ConfiguracionZyrion{Estado: "incierto", Estrategia: "evaluar"},
	}
	resultado, err := motor.Nuevo().EjecutarManifiesto(documento)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.Decision != "evaluar_incierto" {
		t.Fatalf("decisión inesperada: %s", resultado.Decision)
	}
}

func TestMotorCambiaRootCIDCuandoCambiaLaDefinicion(t *testing.T) {
	base := manifiesto.Manifiesto{
		Nombre: "organismo-identidad",
		Zyrion: manifiesto.ConfiguracionZyrion{Estado: "si"},
	}
	alterado := base
	alterado.Capacidades = []string{"gene.ejecutar"}

	primero, err := motor.Nuevo().EjecutarManifiesto(base)
	if err != nil {
		t.Fatal(err)
	}
	segundo, err := motor.Nuevo().EjecutarManifiesto(alterado)
	if err != nil {
		t.Fatal(err)
	}
	if primero.RootCID == segundo.RootCID {
		t.Fatalf("RootCID no cambió al cambiar la definición")
	}
}
