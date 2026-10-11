package organismo

import (
	"testing"

	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestCicloFelizVerificado(t *testing.T) {
	o := Nuevo("demo")
	def := []byte("demo-def")
	cid := rootcid.CrearContenido(def)
	if err := o.Preparar(cid); err != nil {
		t.Fatal(err)
	}
	// Solo presencia: aún no puede ejecutar
	if err := o.Iniciar(); err == nil {
		t.Fatal("sin verificación no debe ejecutar")
	}
	if err := o.VerificarDefinicion(cid, rootcid.VerificarOK(def, cid)); err != nil {
		t.Fatal(err)
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
}

func TestNoEjecutarSinRootCID(t *testing.T) {
	o := Nuevo("x")
	if err := o.Transicionar(Listo); err != nil {
		t.Fatal(err)
	}
	if err := o.Iniciar(); err == nil {
		t.Fatal("debía fallar sin RootCID")
	}
}

func TestTransicionIlegal(t *testing.T) {
	o := Nuevo("y")
	if err := o.Transicionar(Ejecutando); err == nil {
		t.Fatal("creado→ejecutando debe ser ilegal")
	}
}

func TestMemoriaSistemaYaNoBypass(t *testing.T) {
	o := Nuevo("falso-sistema")
	o.Memoria["sistema"] = "1"
	_ = o.Transicionar(Listo)
	if err := o.Iniciar(); err == nil {
		t.Fatal("Memoria sistema no debe eludir RootCID/verificación")
	}
}

func TestFundacionalEjecutaSinRootCIDApp(t *testing.T) {
	o := NuevoFundacional("Maestro", "boot")
	if err := o.Transicionar(Listo); err != nil {
		t.Fatal(err)
	}
	if err := o.Iniciar(); err != nil {
		t.Fatal(err)
	}
	if !o.Fundacional || o.Espacio != EspacioKernel {
		t.Fatal("sello fundacional")
	}
}

func TestRootCIDPresentePeroNoVerificado(t *testing.T) {
	o := Nuevo("z")
	_ = o.Preparar("rootcid:cadena-no-vacia-pero-falsa")
	if err := o.Iniciar(); err == nil {
		t.Fatal("presencia ≠ verificación")
	}
}

func TestFallarYRecuperar(t *testing.T) {
	o := Nuevo("r")
	def := []byte("r")
	cid := rootcid.CrearContenido(def)
	_ = o.Preparar(cid)
	_ = o.VerificarDefinicion(cid, true)
	_ = o.Iniciar()
	if err := o.Fallar("timeout"); err != nil {
		t.Fatal(err)
	}
	if o.Estado != Error {
		t.Fatal(o.Estado)
	}
	if err := o.Transicionar(Listo); err != nil {
		t.Fatal(err)
	}
}
