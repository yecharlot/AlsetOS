package pruebas

import (
    "testing"

    "github.com/yecharlot/AlsetOS/zyrion"
)

func TestZyrionTextos(t *testing.T) {
    casos := map[zyrion.Valor]string{
        zyrion.No: "no",
        zyrion.Si: "si",
        zyrion.Incierto: "incierto",
    }
    for valor, esperado := range casos {
        if valor.Texto() != esperado {
            t.Fatalf("valor %d: esperado %q, obtenido %q", valor, esperado, valor.Texto())
        }
    }
}
