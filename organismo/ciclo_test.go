package organismo

import "testing"

func TestCicloFeliz(t *testing.T) {
	o := Nuevo("demo")
	if o.Estado != Creado {
		t.Fatalf("esperado creado, got %s", o.Estado)
	}
	if err := o.Preparar("bafy-test-cid"); err != nil {
		t.Fatal(err)
	}
	if o.Estado != Listo || o.RootCID != "bafy-test-cid" {
		t.Fatalf("preparar: estado=%s cid=%s", o.Estado, o.RootCID)
	}
	if err := o.Iniciar(); err != nil {
		t.Fatal(err)
	}
	if o.Estado != Ejecutando {
		t.Fatalf("esperado ejecutando")
	}
	if err := o.Detener(); err != nil {
		t.Fatal(err)
	}
	if o.Estado != Detenido {
		t.Fatalf("esperado detenido")
	}
}

func TestNoEjecutarSinRootCID(t *testing.T) {
	o := Nuevo("x")
	// forzar listo ilegalmente no: Transicionar a listo sin cid
	if err := o.Transicionar(Listo); err != nil {
		t.Fatal(err)
	}
	if err := o.Iniciar(); err == nil {
		t.Fatal("debía fallar sin RootCID")
	}
	if o.Estado == Ejecutando {
		t.Fatal("no debió quedar ejecutando")
	}
}

func TestTransicionIlegal(t *testing.T) {
	o := Nuevo("y")
	if err := o.Transicionar(Ejecutando); err == nil {
		t.Fatal("creado→ejecutando debe ser ilegal")
	}
}

func TestSistemaSinRootCID(t *testing.T) {
	o := Nuevo("kernel")
	o.Memoria["sistema"] = "1"
	if err := o.Transicionar(Listo); err != nil {
		t.Fatal(err)
	}
	if err := o.Iniciar(); err != nil {
		t.Fatal(err)
	}
}

func TestCapacidad(t *testing.T) {
	o := Nuevo("z")
	if o.TieneCapacidad("gene.ejecutar") {
		t.Fatal("default deny")
	}
	o.OtorgarCapacidad("gene.ejecutar")
	if !o.TieneCapacidad("gene.ejecutar") {
		t.Fatal("capacidad no otorgada")
	}
}

func TestFallarYRecuperar(t *testing.T) {
	o := Nuevo("r")
	_ = o.Preparar("cid")
	_ = o.Iniciar()
	if err := o.Fallar("timeout"); err != nil {
		t.Fatal(err)
	}
	if o.Estado != Error || o.Memoria["ultimo_error"] != "timeout" {
		t.Fatalf("error no registrado")
	}
	if err := o.Transicionar(Listo); err != nil {
		t.Fatal(err)
	}
}
