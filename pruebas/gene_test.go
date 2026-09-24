package pruebas

import (
    "testing"

    "github.com/yecharlot/AlsetOS/gene"
    "github.com/yecharlot/AlsetOS/organismo"
)

func TestGeneEjecutaYRegistra(t *testing.T) {
    entidad := organismo.Nuevo("prueba")
    gen := gene.Gene{Nombre: "gene-prueba"}

    resultado := gen.Ejecutar(entidad)

    esperado := "Gene gene-prueba ejecutado para prueba"
    if resultado != esperado {
        t.Fatalf("resultado inesperado: %s", resultado)
    }
    if entidad.Memoria["ultimo_resultado"] != esperado {
        t.Fatalf("memoria no registrada: %s", entidad.Memoria["ultimo_resultado"])
    }
}
