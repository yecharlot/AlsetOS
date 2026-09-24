package pruebas

import (
    "testing"

    "github.com/yecharlot/AlsetOS/mind"
    "github.com/yecharlot/AlsetOS/organismo"
    "github.com/yecharlot/AlsetOS/zyrion"
)

func TestMindDecideEjecutar(t *testing.T) {
    entidad := organismo.Nuevo("prueba")
    entidad.Capacidad["gene.ejecutar"] = true
    mente := mind.Mente{}

    if decision := mente.Decidir(zyrion.Si, entidad); decision != "ejecutar_gene" {
        t.Fatalf("decisión inesperada: %s", decision)
    }
}

func TestMindDetieneSinCapacidad(t *testing.T) {
    entidad := organismo.Nuevo("prueba")
    mente := mind.Mente{}

    if decision := mente.Decidir(zyrion.Si, entidad); decision != "detener" {
        t.Fatalf("decisión inesperada: %s", decision)
    }
}
