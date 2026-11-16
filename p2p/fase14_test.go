package p2p

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestPoliticasRetencionPersistenAutorizacionesYCapacidad(t *testing.T) {
	id, err := identidad.Nueva(); if err != nil { t.Fatal(err) }
	path := filepath.Join(t.TempDir(), "leases.json")
	g, err := NuevoGestorLeases(path); if err != nil { t.Fatal(err) }
	peerID := peer.ID(peerIDDeClave(id.ClavePublica))
	if err := g.AutorizarPeer(peerID, 4096); err != nil { t.Fatal(err) }
	if err := g.ConfigurarCapacidad(8192); err != nil { t.Fatal(err) }
	reabierto, err := NuevoGestorLeases(path); if err != nil { t.Fatal(err) }
	if reabierto.permitidos[peerID.String()] != 4096 || reabierto.capacidad != 8192 {
		t.Fatalf("la política no sobrevivió al reinicio: cuota=%d capacidad=%d", reabierto.permitidos[peerID.String()], reabierto.capacidad)
	}
	if err := reabierto.RevocarPeer(peerID); err != nil { t.Fatal(err) }
	final, err := NuevoGestorLeases(path); if err != nil { t.Fatal(err) }
	if _, ok := final.permitidos[peerID.String()]; ok { t.Fatal("la revocación no persistió") }
}

func TestPoliticasRetencionCorruptasFallenCerradas(t *testing.T) {
	path := filepath.Join(t.TempDir(), "leases.json")
	if _, err := NuevoGestorLeases(path); err != nil { t.Fatal(err) }
	if err := os.WriteFile(path+".politicas.json", []byte(`{"schema":"incorrecto"}`), 0600); err != nil { t.Fatal(err) }
	if _, err := NuevoGestorLeases(path); err == nil { t.Fatal("se aceptó una política persistida corrupta") }
}

func TestMedirCapacidadAlmacenamiento(t *testing.T) {
	dir := t.TempDir()
	medicion, err := medirSistemaArchivos(dir)
	if err != nil { t.Fatal(err) }
	if medicion.TotalBytes <= 0 || medicion.DisponibleBytes <= 0 || medicion.ReservableBytes <= 0 || medicion.ReservableBytes > medicion.DisponibleBytes {
		t.Fatalf("medición inválida: %+v", medicion)
	}
}

func TestInventarioReferenciasNoMarcaVersionesComoHuerfanas(t *testing.T) {
	id, err := identidad.Nueva(); if err != nil { t.Fatal(err) }
	almacen, err := NuevoAlmacenContenido(""); if err != nil { t.Fatal(err) }
	contenido := []byte("version referenciada")
	root := rootcid.CrearContenido(contenido)
	if err := almacen.Guardar(root, contenido); err != nil { t.Fatal(err) }
	suelto := []byte("bloque sin descriptor")
	rootSuelto := rootcid.CrearContenido(suelto)
	if err := almacen.Guardar(rootSuelto, suelto); err != nil { t.Fatal(err) }
	indice, err := NuevoIndiceVersiones(""); if err != nil { t.Fatal(err) }
	d, err := CrearDescriptorVersion(id, "organismo:inventario", 1, root, ""); if err != nil { t.Fatal(err) }
	if _, err := indice.Registrar(d); err != nil { t.Fatal(err) }
	n := &Nodo{almacen:almacen, versiones:indice, organismos:map[string][]byte{}}
	items, err := n.InventarioReferencias(); if err != nil { t.Fatal(err) }
	if len(items) != 2 { t.Fatalf("inventario incompleto: %+v", items) }
	for _, item := range items {
		if item.RootCID == root && (!item.Integro || !item.Referenciado || item.CandidatoSinReferencia) { t.Fatalf("versión referenciada mal clasificada: %+v", item) }
		if item.RootCID == rootSuelto && (!item.Presente || item.Referenciado || !item.CandidatoSinReferencia) { t.Fatalf("bloque suelto mal clasificado: %+v", item) }
	}
}
