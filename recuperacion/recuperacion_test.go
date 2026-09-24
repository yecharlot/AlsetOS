package recuperacion

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/p2p"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestEjecutarOrganismoRemoto(t *testing.T) {
	identidadA, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	identidadB, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }

	nodoA, err := p2p.Nuevo(identidadA, "/ip4/127.0.0.1/tcp/0")
	if err != nil { t.Fatal(err) }
	defer nodoA.Cerrar()

	nodoB, err := p2p.Nuevo(identidadB, "/ip4/127.0.0.1/tcp/0")
	if err != nil { t.Fatal(err) }
	defer nodoB.Cerrar()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := nodoA.Host.Connect(ctx, peer.AddrInfo{ID: nodoB.ID(), Addrs: nodoB.Host.Addrs()}); err != nil {
		t.Fatal(err)
	}
	if err := nodoA.BootstrapDHT(ctx); err != nil { t.Fatal(err) }
	if err := nodoB.BootstrapDHT(ctx); err != nil { t.Fatal(err) }

	manifiesto := []byte(`{"nombre":"organismo-recuperado","zyrion":{"estado":"si"},"capacidades":["gene.ejecutar"],"genes":["gene-saludo"]}`)
	root := rootcid.CrearContenido(manifiesto)
	if err := nodoA.AnunciarOrganismo(ctx, root, manifiesto); err != nil { t.Fatal(err) }

	resultado, err := Nuevo(nodoB).EjecutarRemoto(ctx, root)
	if err != nil { t.Fatal(err) }
	if resultado.RootCID != root { t.Fatalf("RootCID distinto: %s", resultado.RootCID) }
	if len(resultado.GenesEjecutados) != 1 { t.Fatalf("genes ejecutados=%d", len(resultado.GenesEjecutados)) }
}
