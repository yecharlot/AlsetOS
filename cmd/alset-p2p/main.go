package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/p2p"
	"github.com/yecharlot/AlsetOS/pulso"
)

func cargarIdentidad(ruta string) (*identidad.Identidad, error) {
	if _, err := os.Stat(ruta); err == nil {
		return identidad.Cargar(ruta)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	nueva, err := identidad.Nueva()
	if err != nil { return nil, err }
	if err := nueva.Guardar(ruta); err != nil { return nil, err }
	return nueva, nil
}

func main() {
	escuchar := flag.String("escuchar", "/ip4/127.0.0.1/tcp/0", "dirección libp2p")
	rutaIdentidad := flag.String("identidad", "estado/identidad-p2p.json", "identidad Ed25519")
	flag.Parse()

	identidadNodo, err := cargarIdentidad(*rutaIdentidad)
	if err != nil { fmt.Printf("error identidad: %v\n", err); os.Exit(1) }

	nodo, err := p2p.Nuevo(identidadNodo, *escuchar)
	if err != nil { fmt.Printf("error P2P: %v\n", err); os.Exit(1) }
	defer nodo.Cerrar()

	fmt.Printf("[P2P-NODE] node-id=%s\n", identidadNodo.ID)
	fmt.Printf("[P2P-PEER] peer-id=%s\n", nodo.ID())
	for _, direccion := range nodo.Direcciones() {
		fmt.Printf("[P2P-ADDR] %s\n", direccion)
	}
	fmt.Println("[P2P] esperando descubrimiento mDNS...")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	fin := time.NewTimer(30 * time.Second)
	defer fin.Stop()

	for {
		select {
		case <-fin.C:
			fmt.Println("[P2P] fin de prueba")
			return
		case <-ticker.C:
			for _, info := range nodo.Conocidos() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				evento := pulso.Pulso{
					Tipo: "descubrimiento",
					Origen: identidadNodo.ID,
					Contenido: "hola-desde-alsetos",
					Fecha: time.Now().UTC(),
				}
				err := nodo.EnviarPulso(ctx, info.ID, evento)
				cancel()
				if err == nil {
					fmt.Printf("[P2P-ENVIADO] destino=%s\n", info.ID)
				}
			}
		}
	}
}
