package organismo

import "testing"

func TestNuevoEstadoCreado(t *testing.T) {
	o := Nuevo("demo")
	if o.Nombre != "demo" || o.Estado != Creado {
		t.Fatalf("%+v", o)
	}
	if o.Memoria == nil || o.Capacidad == nil {
		t.Fatal("maps nil")
	}
}
