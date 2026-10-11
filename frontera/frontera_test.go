package frontera

import (
	"errors"
	"testing"

	"github.com/yecharlot/AlsetOS/organismo"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func orgUsuarioVerificado(nombre string) *organismo.Organismo {
	o := organismo.Nuevo(nombre)
	def := []byte("app:" + nombre + ":v1")
	cid := rootcid.CrearContenido(def)
	_ = o.Preparar(cid)
	_ = o.VerificarDefinicion(cid, rootcid.VerificarOK(def, cid))
	return o
}

func TestUsuarioOKKernelVivo(t *testing.T) {
	rt := NuevoRuntime()
	o := orgUsuarioVerificado("app1")
	if err := rt.Registrar(o); err != nil {
		t.Fatal(err)
	}
	res := rt.EjecutarUsuario("app1", func(o *organismo.Organismo) error {
		o.Memoria["hola"] = "mundo"
		return nil
	})
	if res.Error != "" || res.Panico {
		t.Fatalf("inesperado: %+v", res)
	}
	if !res.KernelVivo || !rt.Vivo() {
		t.Fatal("kernel debía seguir vivo")
	}
}

func TestPanicoUsuarioNoTumbaKernel(t *testing.T) {
	rt := NuevoRuntime()
	o := orgUsuarioVerificado("bomba")
	_ = rt.Registrar(o)
	res := rt.EjecutarUsuario("bomba", func(o *organismo.Organismo) error {
		panic("fallo de app de usuario")
	})
	if !res.Panico {
		t.Fatal("debía capturar pánico")
	}
	if !res.KernelVivo || !rt.Vivo() {
		t.Fatal("el pánico de usuario no debe derribar el kernel lógico")
	}
	if o.Estado != organismo.Error {
		t.Fatalf("organismo debía quedar en error, got %s", o.Estado)
	}
	if rt.FallosUsuario() < 1 {
		t.Fatal("debía contabilizar fallo")
	}
	// segunda app sigue pudiendo correr
	o2 := orgUsuarioVerificado("sana")
	_ = rt.Registrar(o2)
	res2 := rt.EjecutarUsuario("sana", func(o *organismo.Organismo) error { return nil })
	if res2.Error != "" || !rt.Vivo() {
		t.Fatalf("kernel debía atender otra app: %+v", res2)
	}
}

func TestErrorUsuario(t *testing.T) {
	rt := NuevoRuntime()
	o := orgUsuarioVerificado("e")
	_ = rt.Registrar(o)
	res := rt.EjecutarUsuario("e", func(o *organismo.Organismo) error {
		return errors.New("negocio falló")
	})
	if res.Error == "" || o.Estado != organismo.Error {
		t.Fatalf("esperado error de negocio: %+v estado=%s", res, o.Estado)
	}
	if !rt.Vivo() {
		t.Fatal("kernel vivo")
	}
}

func TestSinVerificacionNoEjecuta(t *testing.T) {
	rt := NuevoRuntime()
	o := organismo.Nuevo("x")
	_ = o.Preparar("rootcid:falso")
	// sin VerificarDefinicion
	_ = rt.Registrar(o)
	res := rt.EjecutarUsuario("x", func(o *organismo.Organismo) error { return nil })
	if res.Error == "" {
		t.Fatal("debía bloquear sin definición verificada")
	}
}

func TestFundacionalNoEnFronteraUsuario(t *testing.T) {
	rt := NuevoRuntime()
	f := organismo.NuevoFundacional("Maestro", "boot")
	if err := rt.Registrar(f); err == nil {
		t.Fatal("no registrar fundacionales como usuario")
	}
}
