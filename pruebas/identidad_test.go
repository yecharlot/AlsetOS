package pruebas

import (
	"testing"

	"github.com/yecharlot/AlsetOS/identidad"
)

func TestIdentidadFirmaYVerificacion(t *testing.T) {
	primera, err := identidad.Nueva()
	if err != nil {
		t.Fatal(err)
	}
	segunda, err := identidad.Nueva()
	if err != nil {
		t.Fatal(err)
	}

	mensaje := []byte("pulse:ejecutar_gene")
	firma := primera.Firmar(mensaje)

	if !identidad.Verificar(primera.ClavePublica, mensaje, firma) {
		t.Fatal("firma válida rechazada")
	}
	if identidad.Verificar(segunda.ClavePublica, mensaje, firma) {
		t.Fatal("firma de otro nodo aceptada")
	}
}
