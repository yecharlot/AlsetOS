package pruebas

import (
    "testing"

    "github.com/yecharlot/AlsetOS/lispai"
    "github.com/yecharlot/AlsetOS/organismo"
)

func TestLispAIActualizaMemoria(t *testing.T) {
    entidad := organismo.Nuevo("prueba")
    interprete := lispai.Interprete{}

    resultado := interprete.Evaluar("(recordar origen lisPai)", entidad)

    if resultado != "memoria actualizada" {
        t.Fatalf("resultado inesperado: %s", resultado)
    }
    if entidad.Memoria["origen"] != "lisPai" {
        t.Fatalf("memoria inesperada: %q", entidad.Memoria["origen"])
    }
}
