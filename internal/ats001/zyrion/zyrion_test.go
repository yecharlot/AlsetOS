package zyrion

import "testing"

func TestContratoNumericoATS001(t *testing.T) {
	if int(F) != 0 || int(V) != 1 || int(I) != 2 {
		t.Fatalf("contrato numérico incorrecto: F=%d V=%d I=%d", F, V, I)
	}
	if Eval(false) != F || Eval(true) != V || Eval(nil) != I {
		t.Fatal("Eval no respeta el contrato 0=False, 1=True, 2=Indeterminado")
	}
}
