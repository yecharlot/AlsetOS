package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func esperarBloqueYEstado(t *testing.T, nodo *Nodo, organismo string, roots []string, estado EstadoLinaje) {
	t.Helper()
	limite := time.Now().Add(12 * time.Second)
	for time.Now().Before(limite) {
		if nodo.EstadoLinaje(organismo) == estado {
			completos := true
			for _, root := range roots {
				disponible, err := nodo.DisponibleLocal(root)
				if err != nil || !disponible { completos = false; break }
			}
			if completos { return }
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("el nodo no sincronizó organismo=%s estado=%s esperado=%s", organismo, nodo.EstadoLinaje(organismo), estado)
}

func TestSincronizacionVersionesAlReconectarRecuperaRootCID(t *testing.T) {
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

	organismo := "organismo:reconexion:inventario"
	base := []byte("definicion base")
	rootBase := rootcid.CrearContenido(base)
	descriptorBase, err := CrearDescriptorVersion(idA, organismo, 1, rootBase, "")
	if err != nil { t.Fatal(err) }
	if err := nodoA.almacen.Guardar(rootBase, base); err != nil { t.Fatal(err) }
	if _, err := nodoA.versiones.Registrar(descriptorBase); err != nil { t.Fatal(err) }

	rama := []byte("definicion revision dos")
	rootRama := rootcid.CrearContenido(rama)
	descriptorRama, err := CrearDescriptorVersion(idA, organismo, 2, rootRama, rootBase)
	if err != nil { t.Fatal(err) }
	if err := nodoA.almacen.Guardar(rootRama, rama); err != nil { t.Fatal(err) }
	if _, err := nodoA.versiones.Registrar(descriptorRama); err != nil { t.Fatal(err) }

	// El nodo B permanece desconectado mientras A registra ambas revisiones.
	ctx, cancelar := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelar()
	if err := nodoB.Host.Connect(ctx, peer.AddrInfo{ID:nodoA.ID(), Addrs:nodoA.Host.Addrs()}); err != nil {
		t.Fatal(err)
	}
	esperarBloqueYEstado(t, nodoB, organismo, []string{rootBase, rootRama}, LinajeCoherente)
	versiones := nodoB.Versiones(organismo)
	if len(versiones) != 2 {
		t.Fatalf("se esperaban dos descriptores sincronizados, llegaron %d", len(versiones))
	}
}

func TestSincronizacionConservaForksSinElegirGanador(t *testing.T) {
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

	organismo := "organismo:reconexion:fork"
	uno := []byte("rama A")
	dos := []byte("rama B")
	rootUno, rootDos := rootcid.CrearContenido(uno), rootcid.CrearContenido(dos)
	d1, err := CrearDescriptorVersion(idA, organismo, 1, rootUno, "")
	if err != nil { t.Fatal(err) }
	d2, err := CrearDescriptorVersion(idB, organismo, 1, rootDos, "")
	if err != nil { t.Fatal(err) }
	for _, item := range []struct{root string; data []byte; d DescriptorVersion}{{rootUno,uno,d1},{rootDos,dos,d2}} {
		if err := nodoA.almacen.Guardar(item.root,item.data); err != nil { t.Fatal(err) }
		if _, err := nodoA.versiones.Registrar(item.d); err != nil { t.Fatal(err) }
	}
	ctx, cancelar := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelar()
	if err := nodoB.Host.Connect(ctx, peer.AddrInfo{ID:nodoA.ID(), Addrs:nodoA.Host.Addrs()}); err != nil { t.Fatal(err) }
	esperarBloqueYEstado(t, nodoB, organismo, []string{rootUno,rootDos}, LinajeConflicto)
	if got := len(nodoB.Versiones(organismo)); got != 2 {
		t.Fatalf("la sincronización descartó una rama; descriptores=%d", got)
	}
}
