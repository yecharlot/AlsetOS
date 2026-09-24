package sandbox

import (
    "testing"

    "github.com/yecharlot/AlsetOS/zyrion"
)

func TestSimulacionEjecutaConSi(t *testing.T) {
    resultado := EjecutarSimulacion("prueba-si", zyrion.Si, true)
    if !resultado.Ejecutado {
        t.Fatal("el Gene debía ejecutarse")
    }
    if resultado.Decision != "ejecutar_gene" {
        t.Fatalf("decisión inesperada: %s", resultado.Decision)
    }
}

func TestSimulacionDetieneConIncierto(t *testing.T) {
    resultado := EjecutarSimulacion("prueba-incierto", zyrion.Incierto, true)
    if resultado.Ejecutado {
        t.Fatal("el Gene no debía ejecutarse")
    }
    if resultado.Decision != "detener" {
        t.Fatalf("decisión inesperada: %s", resultado.Decision)
    }
}
