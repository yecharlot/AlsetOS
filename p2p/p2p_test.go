package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/pulso"
)

func TestPulseEntreNodos(t *testing.T) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := nodoA.Host.Connect(ctx, peer.AddrInfo{ID:nodoB.ID(), Addrs:nodoB.Host.Addrs()}); err != nil {
		t.Fatal(err)
	}

	evento := pulso.Pulso{Tipo:"prueba", Origen:identidadA.ID, Contenido:"hola-nodo-b", Fecha:time.Now().UTC()}
	if err := nodoA.EnviarPulso(ctx, nodoB.ID(), evento); err != nil { t.Fatal(err) }
}
