package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/yecharlot/AlsetOS/autonomia"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/p2p"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func cargarIdentidad(ruta string) (*identidad.Identidad, error) {
	if _, err := os.Stat(ruta); err == nil {
		return identidad.Cargar(ruta)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	nueva, err := identidad.Nueva()
	if err != nil {
		return nil, err
	}
	if err := nueva.Guardar(ruta); err != nil {
		return nil, err
	}
	return nueva, nil
}

func conectar(ctx context.Context, nodo *p2p.Nodo, direccion string) error {
	if direccion == "" {
		return nil
	}
	multi, err := multiaddr.NewMultiaddr(direccion)
	if err != nil {
		return err
	}
	info, err := peer.AddrInfoFromP2pAddr(multi)
	if err != nil {
		return err
	}
	return nodo.Host.Connect(ctx, *info)
}

func main() {
	escuchar := flag.String("escuchar", "/ip4/127.0.0.1/tcp/0", "dirección libp2p")
	peerRemoto := flag.String("peer", "", "multiaddr /p2p del nodo remoto")
	identidadRuta := flag.String("identidad", "estado/identidad-p2p.json", "identidad persistente")
	organismosRuta := flag.String("organismos", "estado/organismos.json", "almacén de organismos")
	colocacionRuta := flag.String("colocacion", "estado/colocacion.json", "registro de placement")
	publicar := flag.String("publicar", "", "manifiesto .alset a publicar y replicar")
	replicas := flag.Int("replicas", 2, "cantidad de réplicas deseadas")
	duracion := flag.Duration("duracion", 30*time.Second, "duración del monitor")
	flag.Parse()

	identidadNodo, err := cargarIdentidad(*identidadRuta)
	if err != nil {
		fmt.Printf("error identidad: %v\n", err)
		os.Exit(1)
	}
	nodo, err := p2p.NuevoConEstado(identidadNodo, *escuchar, *organismosRuta)
	if err != nil {
		fmt.Printf("error nodo: %v\n", err)
		os.Exit(1)
	}
	defer nodo.Cerrar()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := conectar(ctx, nodo, *peerRemoto); err != nil {
		fmt.Printf("error peer: %v\n", err)
		os.Exit(1)
	}
	if *peerRemoto != "" {
		if err := nodo.BootstrapDHT(ctx); err != nil {
			fmt.Printf("error DHT: %v\n", err)
			os.Exit(1)
		}
	}

	servicio, err := autonomia.Nuevo(nodo, *colocacionRuta)
	if err != nil {
		fmt.Printf("error autonomía: %v\n", err)
		os.Exit(1)
	}
	servicio.FactorReplica = *replicas

	if *publicar != "" {
		documento, err := manifiesto.Cargar(*publicar)
		if err != nil {
			fmt.Printf("error manifiesto: %v\n", err)
			os.Exit(1)
		}
		contenido, err := documento.Canonico()
		if err != nil {
			fmt.Printf("error canónico: %v\n", err)
			os.Exit(1)
		}
		root := rootcid.CrearContenido(contenido)
		colocacion, err := servicio.Publicar(ctx, root, contenido)
		if err != nil {
			fmt.Printf("error publicar: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[ALSET-PLACEMENT] rootcid=%s primario=%s replicas=%v\n", colocacion.RootCID, colocacion.Primario, colocacion.Replicas)
	}

	fmt.Printf("[ALSET-AUTONOMIA] node=%s intervalo=%s timeout=%s\n", nodo.ID(), servicio.Intervalo, servicio.Timeout)
	monitorCtx, monitorCancel := context.WithCancel(context.Background())
	defer monitorCancel()
	go servicio.Ejecutar(monitorCtx)

	timer := time.NewTimer(*duracion)
	defer timer.Stop()
	<-timer.C

	for _, colocacion := range servicio.Registro.Todos() {
		fmt.Printf("[ALSET-PLACEMENT] rootcid=%s primario=%s replicas=%v\n", colocacion.RootCID, colocacion.Primario, colocacion.Replicas)
	}
}
