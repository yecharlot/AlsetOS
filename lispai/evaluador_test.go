package lispai

import (
	"testing"

	"github.com/yecharlot/AlsetOS/organismo"
)

func TestLispAICondicional(t *testing.T) {
	e := organismo.Nuevo("lisp")
	if _, err := (Interprete{}).Ejecutar("(recordar estado listo)", e); err != nil {
		t.Fatal(err)
	}
	r, err := (Interprete{}).Ejecutar("(si (igual (leer estado) listo) (recordar decision ejecutar))", e)
	if err != nil {
		t.Fatal(err)
	}
	if r.Valor != "ejecutar" || e.Memoria["decision"] != "ejecutar" {
		t.Fatalf("resultado inesperado: %#v %#v", r, e.Memoria)
	}
}
