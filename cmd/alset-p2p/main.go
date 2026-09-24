package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/p2p"
	"github.com/yecharlot/AlsetOS/pulso"
	"github.com/yecharlot/AlsetOS/recuperacion"
	"github.com/yecharlot/AlsetOS/rootcid"
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

func conectar(ctx context.Context, nodo *p2p.Nodo, direccion string) error {
	if direccion == "" { return nil }
	multi, err := multiaddr.NewMultiaddr(direccion)
	if err != nil { return fmt.Errorf("dirección peer inválida: %w", err) }
	info, err := peer.AddrInfoFromP2pAddr(multi)
	if err != nil { return fmt.Errorf("interpretar peer: %w", err) }
	if err := nodo.Host.Connect(ctx, *info); err != nil {
		return fmt.Errorf("conectar peer: %w", err)
	}
	return nil
}

func main() {
	escuchar := flag.String("escuchar", "/ip4/127.0.0.1/tcp/0", "dirección libp2p")
	rutaIdentidad := flag.String("identidad", "estado/identidad-p2p.json", "identidad Ed25519")
	rutaOrganismos := flag.String("organismos", "estado/organismos.json", "almacén persistente de organismos")
	peerRemoto := flag.String("peer", "", "multiaddr /p2p del nodo remoto")
	anunciar := flag.String("anunciar", "", "manifiesto .alset que se anunciará en la DHT")
	recuperar := flag.String("recuperar", "", "RootCID que se buscará y ejecutará remotamente")
	espera := flag.Duration("espera", 30*time.Second, "tiempo de ejecución de la prueba")
	flag.Parse()

	identidadNodo, err := cargarIdentidad(*rutaIdentidad)
	if err != nil { fmt.Printf("error identidad: %v\n", err); os.Exit(1) }

	nodo, err := p2p.NuevoConEstado(identidadNodo, *escuchar, *rutaOrganismos)
	if err != nil { fmt.Printf("error P2P: %v\n", err); os.Exit(1) }
	defer nodo.Cerrar()

	fmt.Printf("[P2P-NODE] node-id=%s\n", identidadNodo.ID)
	fmt.Printf("[P2P-PEER] peer-id=%s\n", nodo.ID())
	for _, direccion := range nodo.Direcciones() {
		fmt.Printf("[P2P-ADDR] %s\n", direccion)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := conectar(ctx, nodo, *peerRemoto); err != nil {
		fmt.Printf("error conexión: %v\n", err)
		os.Exit(1)
	}
	if *peerRemoto != "" {
		if err := nodo.BootstrapDHT(ctx); err != nil {
			fmt.Printf("error bootstrap DHT: %v\n", err)
			os.Exit(1)
		}
		if err := nodo.ReanunciarOrganismos(ctx); err != nil {
			fmt.Printf("error reanunciar organismos: %v\n", err)
			os.Exit(1)
		}
	}

	if *anunciar != "" {
		documento, err := manifiesto.Cargar(*anunciar)
		if err != nil { fmt.Printf("error manifiesto: %v\n", err); os.Exit(1) }
		contenido, err := documento.Canonico()
		if err != nil { fmt.Printf("error canónico: %v\n", err); os.Exit(1) }
		root := rootcid.CrearContenido(contenido)
		if err := nodo.AnunciarOrganismo(ctx, root, contenido); err != nil {
			fmt.Printf("error anuncio: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[DHT-ANUNCIADO] rootcid=%s\n", root)
	}

	if *recuperar != "" {
		resultado, err := recuperacion.Nuevo(nodo).EjecutarRemoto(ctx, *recuperar)
		if err != nil {
			fmt.Printf("error recuperación: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[RECUPERADO] organismo=%s rootcid=%s decisión=%s\n", resultado.Nombre, resultado.RootCID, resultado.Decision)
		for _, gene := range resultado.GenesEjecutados {
			fmt.Printf("[RECUPERADO-GENE] %s\n", gene)
		}
	}

	if *anunciar == "" && *recuperar == "" {
		fmt.Println("[P2P] esperando descubrimiento mDNS...")
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		fin := time.NewTimer(*espera)
		defer fin.Stop()
		for {
			select {
			case <-fin:
				fmt.Println("[P2P] fin de prueba")
				return
			case <-ticker.C:
				for _, info := range nodo.Conocidos() {
					evento := pulso.Pulso{
						Tipo: "descubrimiento",
						Origen: identidadNodo.ID,
						Contenido: "hola-desde-alsetos",
						Fecha: time.Now().UTC(),
					}
					ctxPulso, cancelPulso := context.WithTimeout(context.Background(), 5*time.Second)
					err := nodo.EnviarPulso(ctxPulso, info.ID, evento)
					cancelPulso()
					if err == nil {
						fmt.Printf("[P2P-ENVIADO] destino=%s\n", info.ID)
					}
				}
			}
		}
	}
}
