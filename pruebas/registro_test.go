package pruebas

import (
	"testing"

	"github.com/yecharlot/AlsetOS/registro"
)

func TestRegistroPersisteRootCID(t *testing.T) {
	ruta := t.TempDir() + "/estado/registro.json"

	primero, err := registro.Nuevo(ruta)
	if err != nil { t.Fatal(err) }

	entrada := registro.Entrada{
		Nombre: "organismo",
		RootCID: "rootcid:abc",
		Estado: "detenido",
		Actualizado: "ahora",
	}
	if err := primero.Guardar(entrada); err != nil { t.Fatal(err) }

	segundo, err := registro.Nuevo(ruta)
	if err != nil { t.Fatal(err) }

	recuperada, existe := segundo.Obtener("rootcid:abc")
	if !existe || recuperada.Nombre != "organismo" {
		t.Fatalf("entrada no recuperada: %+v", recuperada)
	}
}
