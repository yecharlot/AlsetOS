package pruebas

import (
    "testing"

    "github.com/yecharlot/AlsetOS/rootcid"
)

func TestRootCIDDeterminista(t *testing.T) {
    primero := rootcid.Crear("organismo-prueba")
    segundo := rootcid.Crear("organismo-prueba")
    if primero != segundo {
        t.Fatalf("RootCID no determinista: %s != %s", primero, segundo)
    }
}
