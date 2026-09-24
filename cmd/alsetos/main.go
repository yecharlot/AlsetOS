package main

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

func EjecutarFlujo() string {
	entidad := organismo.Nuevo("organismo-prueba")
	entidad.RootCID = rootcid.Crear(entidad.Nombre)
	entidad.Capacidad["gene.ejecutar"] = true

	fmt.Printf("[ORGANISMO] %s\n", entidad.Nombre)
	fmt.Printf("[ROOTCID] %s\n", entidad.RootCID)

	interprete := lispai.Interprete{}
	fmt.Printf("[LISPAI] %s\n", interprete.Evaluar("(recordar origen lisPai)", entidad))

	estado := zyrion.Si
	fmt.Printf("[ZYRION] estado=%s\n", estado.Texto())

	mente := mind.Mente{}
	decision := mente.Decidir(estado, entidad)
	fmt.Printf("[MIND] decisión=%s\n", decision)

	gen := gene.Gene{Nombre: "gene-saludo"}
	var resultado string
	if decision == "ejecutar_gene" {
		resultado = gen.Ejecutar(entidad)
		fmt.Printf("[GENE] %s\n", resultado)
	}

	pulso.Emitir(entidad, pulso.Pulso{
		Tipo:      "resultado",
		Origen:    entidad.RootCID,
		Contenido: resultado,
		Fecha:     time.Now(),
	})

	return resultado
}

func main() {
	resultado := EjecutarFlujo()
	if resultado == "" {
		panic("flujo AlsetOS sin resultado")
	}
	fmt.Printf("[RESULTADO] %s\n", resultado)
}
