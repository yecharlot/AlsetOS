package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestDHTEncuentraYRecuperaOrganismo(t *testing.T) {
	identidadA, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	identidadB, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }

	nodoA, err := Nuevo(identidadA, "/ip4/127.0.0.1/tcp/0")
	if err != nil { t.Fatal(err) }
	defer nodoA.Cerrar()

	nodoB, err := Nuevo(identidadB, "/ip4/127.0.0.1/tcp/0")
	if err != nil { t.Fatal(err) }
	defer nodoB.Cerrar()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := nodoA.Host.Connect(ctx, peer.AddrInfo{ID: nodoB.ID(), Addrs: nodoB.Host.Addrs()}); err != nil {
		t.Fatal(err)
	}

	if err := nodoA.BootstrapDHT(ctx); err != nil { t.Fatal(err) }
	if err := nodoB.BootstrapDHT(ctx); err != nil { t.Fatal(err) }

	manifiesto := []byte(`{"nombre":"organismo-remoto","zyrion":{"estado":"si"},"genes":["gene-saludo"]}`)
	root := rootcid.CrearContenido(manifiesto)

	if err := nodoA.AnunciarOrganismo(ctx, root, manifiesto); err != nil {
		t.Fatal(err)
	}

	recuperado, err := nodoB.RecuperarOrganismo(ctx, root)
	if err != nil { t.Fatal(err) }
	if string(recuperado) != string(manifiesto) {
		t.Fatalf("manifiesto distinto: %s", recuperado)
	}
}
