package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/nodo"
)

func main() {
	registro := flag.String("registro", "estado/registro.json", "ruta del registro persistente del nodo")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("uso: go run ./cmd/alset-node [--registro estado/registro.json] <organismo.alset>")
		os.Exit(2)
	}

	rutaManifiesto := flag.Arg(0)
	documento, err := manifiesto.Cargar(rutaManifiesto)
	if err != nil {
		fmt.Printf("error manifiesto: %v\n", err)
		os.Exit(1)
	}

	host, err := nodo.Nuevo("alset-node-local", *registro)
	if err != nil {
		fmt.Printf("error nodo: %v\n", err)
		os.Exit(1)
	}

	resultado, err := host.Ejecutar(documento)
	if err != nil {
		fmt.Printf("error ejecución: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[NODE] %s estado=%s\n", host.Nombre, host.Estado)
	fmt.Printf("[ORGANISMO] %s\n", resultado.Organismo.Nombre)
	fmt.Printf("[ROOTCID] %s\n", resultado.Organismo.RootCID)
	fmt.Printf("[DECISION] %s\n", resultado.Organismo.Decision)
	fmt.Printf("[REGISTRO] %s\n", *registro)
}
