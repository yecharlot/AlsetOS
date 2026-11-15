package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestReplicacionDeVersionFirmadaYConflictoEntreNodos(t *testing.T) {
	idA, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	idB, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	nodoA, err := Nuevo(idA, "/ip4/127.0.0.1/tcp/0")
	if err != nil { t.Fatal(err) }
	defer nodoA.Cerrar()
	nodoB, err := Nuevo(idB, "/ip4/127.0.0.1/tcp/0")
	if err != nil { t.Fatal(err) }
	defer nodoB.Cerrar()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := nodoA.Host.Connect(ctx, peer.AddrInfo{ID:nodoB.ID(), Addrs:nodoB.Host.Addrs()}); err != nil { t.Fatal(err) }

	contenido1 := []byte("definicion original")
	root1 := rootcid.CrearContenido(contenido1)
	_, estado, err := nodoA.AnunciarVersion(ctx, nodoB.ID(), "organismo:almacen", 1, root1, "", contenido1)
	if err != nil || estado != LinajeCoherente { t.Fatalf("replicar v1: estado=%s err=%v", estado, err) }
	if nodoB.EstadoLinaje("organismo:almacen") != LinajeCoherente { t.Fatal("el nodo receptor no registró el linaje inicial") }
	recuperado, err := nodoB.RecuperarOrganismo(ctx, root1)
	if err != nil || string(recuperado) != string(contenido1) { t.Fatalf("el nodo receptor no conserva el bloque: %v", err) }

	contenidoFork := []byte("otra definición para la misma revisión")
	rootFork := rootcid.CrearContenido(contenidoFork)
	fork, err := CrearDescriptorVersion(idB, "organismo:almacen", 1, rootFork, "")
	if err != nil { t.Fatal(err) }
	estado, err = nodoA.ReplicarDescriptor(ctx, nodoB.ID(), fork, contenidoFork)
	if err != nil { t.Fatal(err) }
	if estado != LinajeConflicto || nodoB.EstadoLinaje("organismo:almacen") != LinajeConflicto {
		t.Fatalf("se esperaba conflicto conservado en receptor, estado=%s", estado)
	}
	if len(nodoB.Versiones("organismo:almacen")) != 2 { t.Fatal("el receptor debe conservar las dos versiones en conflicto") }
}
