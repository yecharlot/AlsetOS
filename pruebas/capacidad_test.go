package pruebas

import (
	"testing"
	"github.com/yecharlot/AlsetOS/capacidad"
	"github.com/yecharlot/AlsetOS/organismo"
)

func TestCapacidadRequerida(t *testing.T) {
	entidad := organismo.Nuevo("prueba")
	if err := capacidad.Requerir(entidad, "gene.ejecutar"); err == nil { t.Fatal("debía rechazar capacidad ausente") }
	entidad.Capacidad["gene.ejecutar"] = true
	if err := capacidad.Requerir(entidad, "gene.ejecutar"); err != nil { t.Fatal(err) }
}
