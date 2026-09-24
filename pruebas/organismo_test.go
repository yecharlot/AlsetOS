package pruebas

import (
    "testing"

    "github.com/yecharlot/AlsetOS/organismo"
)

func TestCrearOrganismo(t *testing.T) {
    entidad := organismo.Nuevo("prueba")
    if entidad.Nombre != "prueba" {
        t.Fatalf("nombre inesperado: %s", entidad.Nombre)
    }
    if entidad.Memoria == nil || entidad.Capacidad == nil {
        t.Fatal("memoria y capacidades deben inicializarse")
    }
}
