package sandbox

import (
    "fmt"
    "time"

    "github.com/yecharlot/AlsetOS/gene"
    "github.com/yecharlot/AlsetOS/lispai"
    "github.com/yecharlot/AlsetOS/mind"
    "github.com/yecharlot/AlsetOS/organismo"
    "github.com/yecharlot/AlsetOS/pulso"
    "github.com/yecharlot/AlsetOS/rootcid"
    "github.com/yecharlot/AlsetOS/zyrion"
)

// Resultado representa el estado observable de una simulación.
type Resultado struct {
    Nombre      string
    RootCID     string
    Estado      zyrion.Valor
    Decision    string
    Ejecutado   bool
    UltimoPulso string
}

// EjecutarSimulacion ejecuta un organismo completamente en memoria.
func EjecutarSimulacion(nombre string, estado zyrion.Valor, capacidad bool) Resultado {
    entidad := organismo.Nuevo(nombre)
    entidad.RootCID = rootcid.Crear(nombre)
    entidad.Capacidad["gene.ejecutar"] = capacidad

    interprete := lispai.Interprete{}
    interprete.Evaluar("(recordar entorno sandbox)", entidad)

    mente := mind.Mente{}
    decision := mente.Decidir(estado, entidad)

    ejecutado := false
    if decision == "ejecutar_gene" {
        gen := gene.Gene{Nombre: "gene-sandbox"}
        gen.Ejecutar(entidad)
        ejecutado = true
    }

    pulso.Emitir(entidad, pulso.Pulso{
        Tipo:      "simulacion",
        Origen:    entidad.RootCID,
        Contenido: decision,
        Fecha:      time.Now(),
    })

    return Resultado{
        Nombre:      entidad.Nombre,
        RootCID:     entidad.RootCID,
        Estado:      estado,
        Decision:    decision,
        Ejecutado:   ejecutado,
        UltimoPulso: entidad.Memoria["ultimo_pulso"],
    }
}

// Mostrar imprime una simulación sin depender de red, disco ni hardware.
func Mostrar(resultado Resultado) {
    fmt.Printf("[SANDBOX] organismo=%s\n", resultado.Nombre)
    fmt.Printf("[SANDBOX] rootcid=%s\n", resultado.RootCID)
    fmt.Printf("[SANDBOX] zyrion=%s\n", resultado.Estado.Texto())
    fmt.Printf("[SANDBOX] mind=%s\n", resultado.Decision)
    fmt.Printf("[SANDBOX] gene_ejecutado=%t\n", resultado.Ejecutado)
    fmt.Printf("[SANDBOX] pulso=%s\n", resultado.UltimoPulso)
}
