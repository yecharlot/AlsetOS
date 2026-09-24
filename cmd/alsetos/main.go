package main

import (
	"fmt"
	"os"

	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/motor"
)

func EjecutarManifiesto(ruta string) error {
	documento, err := manifiesto.Cargar(ruta)
	if err != nil { return err }

	resultado, err := motor.Nuevo().EjecutarManifiesto(documento)
	if err != nil { return err }

	fmt.Printf("[ORGANISMO] %s\n", resultado.Nombre)
	fmt.Printf("[ROOTCID] %s\n", resultado.RootCID)
	fmt.Printf("[ZYRION] estado=%s\n", resultado.Estado)
	fmt.Printf("[MIND] decisión=%s\n", resultado.Decision)
	for _, ejecucion := range resultado.GenesEjecutados { fmt.Printf("[GENE] %s\n", ejecucion) }
	fmt.Printf("[PULSO] %s\n", resultado.UltimoPulso)
	fmt.Printf("[RESULTADO] %s\n", resultado.Decision)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("uso: go run ./cmd/alsetos <organismo.alset>")
		fmt.Println("ejemplo: go run ./cmd/alsetos organismos/organismo-si.alset")
		os.Exit(2)
	}
	if err := EjecutarManifiesto(os.Args[1]); err != nil {
		fmt.Printf("error AlsetOS: %v\n", err)
		os.Exit(1)
	}
}
