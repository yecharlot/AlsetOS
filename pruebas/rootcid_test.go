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

func TestRootCIDDependeDelContenido(t *testing.T) {
	primero := rootcid.CrearContenido([]byte("organismo-a"))
	segundo := rootcid.CrearContenido([]byte("organismo-a"))
	tercero := rootcid.CrearContenido([]byte("organismo-b"))

	if primero != segundo {
		t.Fatalf("RootCID por contenido no determinista")
	}
	if primero == tercero {
		t.Fatalf("contenidos distintos produjeron el mismo RootCID")
	}
}
