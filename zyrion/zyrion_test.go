package zyrion

import "testing"

func TestTexto(t *testing.T) {
	if No.Texto() != "no" || Si.Texto() != "si" || Incierto.Texto() != "incierto" {
		t.Fatalf("%s %s %s", No.Texto(), Si.Texto(), Incierto.Texto())
	}
}

func TestValoresDistintos(t *testing.T) {
	if No == Si || Si == Incierto {
		t.Fatal("ternary collapsed")
	}
}
