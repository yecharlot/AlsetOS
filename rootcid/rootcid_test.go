package rootcid

import "testing"

func TestVerificarDistintoDePresente(t *testing.T) {
	c := []byte("def")
	cid := CrearContenido(c)
	if !Presente(cid) {
		t.Fatal("presente")
	}
	if err := Verificar(c, cid); err != nil {
		t.Fatal(err)
	}
	if err := Verificar(c, "rootcid:otra"); err == nil {
		t.Fatal("debe fallar")
	}
	if Presente("x") && VerificarOK(c, "x") {
		t.Fatal("presente no implica verificado")
	}
}
