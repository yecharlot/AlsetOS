package pruebas

import (
    "testing"
    "time"

    "github.com/yecharlot/AlsetOS/organismo"
    "github.com/yecharlot/AlsetOS/pulso"
)

func TestPulsoRegistraEvento(t *testing.T) {
    entidad := organismo.Nuevo("prueba")
    pulso.Emitir(entidad, pulso.Pulso{
        Tipo: "prueba",
        Origen: "origen",
        Contenido: "contenido",
        Fecha: time.Now(),
    })

    if entidad.Memoria["ultimo_pulso"] != "prueba:contenido" {
        t.Fatalf("pulso no registrado: %s", entidad.Memoria["ultimo_pulso"])
    }
}
