package p2p

import (
	"context"
	"errors"
	"testing"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func crearTrioGossip(t *testing.T) ([]*Nodo, context.Context, context.CancelFunc) {
	t.Helper()
	nodos := make([]*Nodo, 0, 3)
	for i := 0; i < 3; i++ {
		id, err := identidad.Nueva()
		if err != nil {
			t.Fatal(err)
		}
		clave, err := libp2pcrypto.UnmarshalEd25519PrivateKey(id.ClavePrivada)
		if err != nil {
			t.Fatal(err)
		}
		h, err := libp2p.New(libp2p.Identity(clave), libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
		if err != nil {
			t.Fatal(err)
		}
		almacen, err := NuevoAlmacenContenido("")
		if err != nil {
			_ = h.Close()
			t.Fatal(err)
		}
		indice, err := NuevoIndiceVersiones("")
		if err != nil {
			_ = h.Close()
			t.Fatal(err)
		}
		nodo := &Nodo{Host: h, Identidad: id, almacen: almacen, versiones: indice}
		if err := nodo.iniciarGossipVersiones(); err != nil {
			_ = h.Close()
			t.Fatal(err)
		}
		nodos = append(nodos, nodo)
	}
	t.Cleanup(func() {
		for _, nodo := range nodos {
			_ = nodo.Cerrar()
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	for i := range nodos {
		for j := i + 1; j < len(nodos); j++ {
			if err := nodos[i].Host.Connect(ctx, peer.AddrInfo{ID: nodos[j].ID(), Addrs: nodos[j].Host.Addrs()}); err != nil {
				t.Fatalf("conectar nodos %d y %d: %v", i, j, err)
			}
		}
	}
	for i, nodo := range nodos {
		if err := nodo.EsperarMallaVersiones(ctx, 2); err != nil {
			t.Fatalf("nodo %d no formó malla GossipSub: %v (peers=%v)", i, err, nodo.topicVersiones.ListPeers())
		}
	}
	return nodos, ctx, cancel
}

func esperarEstadoLinaje(t *testing.T, ctx context.Context, nodos []*Nodo, organismo string, esperado EstadoLinaje, cantidad int) {
	t.Helper()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		completos := 0
		for _, nodo := range nodos {
			if nodo.EstadoLinaje(organismo) == esperado {
				completos++
			}
		}
		if completos >= cantidad {
			return
		}
		select {
		case <-ctx.Done():
			estados := make([]EstadoLinaje, len(nodos))
			for i, nodo := range nodos {
				estados[i] = nodo.EstadoLinaje(organismo)
			}
			t.Fatalf("no se propagó estado %s; estados=%v: %v", esperado, estados, ctx.Err())
		case <-ticker.C:
		}
	}
}

func TestGossipSubPropagaAnuncioFirmadoATresNodos(t *testing.T) {
	nodos, ctx, _ := crearTrioGossip(t)
	contenido := []byte("organismo distribuido")
	root := rootcid.CrearContenido(contenido)
	descriptor, estado, err := nodos[0].AnunciarVersion(ctx, "", "organismo:gossip:trio", 1, root, "", contenido)
	if err != nil {
		t.Fatal(err)
	}
	if estado != LinajeCoherente {
		t.Fatalf("estado local inesperado: %s", estado)
	}
	esperarEstadoLinaje(t, ctx, nodos, "organismo:gossip:trio", LinajeCoherente, 3)
	for i, nodo := range nodos {
		versiones := nodo.Versiones("organismo:gossip:trio")
		if len(versiones) != 1 || versiones[0].RootCID != descriptor.RootCID {
			t.Fatalf("nodo %d no registró exactamente el descriptor anunciado: %+v", i, versiones)
		}
		if i > 0 {
			disponible, err := nodo.almacen.Disponible(root)
			if err != nil || disponible {
				t.Fatalf("nodo %d confundió anuncio con disponibilidad del bloque: disponible=%v err=%v", i, disponible, err)
			}
		}
	}
}

func TestGossipSubRechazaAnuncioMalformadoYFirmaAlterada(t *testing.T) {
	nodos, ctx, _ := crearTrioGossip(t)
	if err := nodos[0].topicVersiones.Publish(ctx, []byte(`{"schema":"incorrecto"}`)); err == nil {
		t.Fatal("GossipSub aceptó un anuncio malformado")
	}

	id, err := identidad.Nueva()
	if err != nil {
		t.Fatal(err)
	}
	d, err := CrearDescriptorVersion(id, "organismo:malicioso", 1, rootcid.CrearContenido([]byte("legitimo")), "")
	if err != nil {
		t.Fatal(err)
	}
	d.RootCID = rootcid.CrearContenido([]byte("alterado"))
	if _, err := codificarAnuncioVersion(d); err == nil {
		t.Fatal("el codificador aceptó un descriptor con firma alterada")
	}
	if validarMensajeAnuncioVersion(ctx, nodos[0].ID(), nil) {
		t.Fatal("el validador aceptó un mensaje nulo")
	}
}

func TestGossipSubOrdenInciertoForkYDeduplicacion(t *testing.T) {
	nodos, ctx, _ := crearTrioGossip(t)
	const organismo = "organismo:gossip:linaje"
	v1 := []byte("base")
	root1 := rootcid.CrearContenido(v1)
	v2 := []byte("rama principal")
	root2 := rootcid.CrearContenido(v2)
	v2Fork := []byte("rama alternativa")
	root2Fork := rootcid.CrearContenido(v2Fork)

	// La revisión 2 llega antes que su antecesor: debe conservarse como incompleta.
	if _, estado, err := nodos[0].AnunciarVersion(ctx, "", organismo, 2, root2, root1, v2); err != nil || estado != LinajeIncompleto {
		t.Fatalf("anuncio fuera de orden: estado=%s err=%v", estado, err)
	}
	esperarEstadoLinaje(t, ctx, nodos, organismo, LinajeIncompleto, 3)

	// La llegada de la revisión 1 completa el linaje sin borrar la revisión 2.
	d1, estado, err := nodos[0].AnunciarVersion(ctx, "", organismo, 1, root1, "", v1)
	if err != nil || estado != LinajeCoherente {
		t.Fatalf("anuncio del antecesor: estado=%s err=%v", estado, err)
	}
	esperarEstadoLinaje(t, ctx, nodos, organismo, LinajeCoherente, 3)

	// Repetir el mismo descriptor no debe duplicarlo en el índice.
	if err := nodos[0].PublicarAnuncioVersion(ctx, d1); err != nil {
		t.Fatal(err)
	}

	// Una segunda rama para la revisión 2 se conserva como conflicto.
	if _, estado, err := nodos[1].AnunciarVersion(ctx, "", organismo, 2, root2Fork, root1, v2Fork); err != nil || estado != LinajeConflicto {
		t.Fatalf("registrar rama alternativa local: estado=%s err=%v", estado, err)
	}
	esperarEstadoLinaje(t, ctx, nodos, organismo, LinajeConflicto, 3)
	for i, nodo := range nodos {
		if got := len(nodo.Versiones(organismo)); got != 3 {
			t.Fatalf("nodo %d debía conservar antecesor y dos ramas, tiene %d descriptores", i, got)
		}
	}
}

func TestGossipSubDescriptorRechazadoSiNoTieneFirmaValida(t *testing.T) {
	id, err := identidad.Nueva()
	if err != nil {
		t.Fatal(err)
	}
	d, err := CrearDescriptorVersion(id, "organismo:invalid", 1, rootcid.CrearContenido([]byte("ok")), "")
	if err != nil {
		t.Fatal(err)
	}
	d.Firma = "firma-invalida"
	if _, err := codificarAnuncioVersion(d); !errors.Is(err, ErrVersionDescriptorInvalido) {
		t.Fatalf("se esperaba descriptor inválido, err=%v", err)
	}
}
