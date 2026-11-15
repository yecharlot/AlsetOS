package p2p

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestAlmacenContenidoPersistenteVerificaYRecupera(t *testing.T) {
	dir := t.TempDir()
	contenido := []byte("manifiesto alset")
	root := rootcid.CrearContenido(contenido)
	almacen, err := NuevoAlmacenContenido(dir)
	if err != nil { t.Fatal(err) }
	if err := almacen.Guardar(root, contenido); err != nil { t.Fatal(err) }
	if err := almacen.Guardar(root, contenido); err != nil { t.Fatalf("guardar idempotente: %v", err) }

	reabierto, err := NuevoAlmacenContenido(dir)
	if err != nil { t.Fatal(err) }
	recuperado, err := reabierto.Obtener(root)
	if err != nil { t.Fatal(err) }
	if string(recuperado) != string(contenido) { t.Fatalf("contenido distinto: %q", recuperado) }
	disponible, err := reabierto.Disponible(root)
	if err != nil || !disponible { t.Fatalf("disponibilidad=%v err=%v", disponible, err) }
}

func TestAlmacenContenidoRechazaCIDIncorrectoYCorrupcion(t *testing.T) {
	dir := t.TempDir()
	contenido := []byte("contenido correcto")
	root := rootcid.CrearContenido(contenido)
	almacen, err := NuevoAlmacenContenido(dir)
	if err != nil { t.Fatal(err) }
	if err := almacen.Guardar(root, []byte("contenido alterado")); err == nil { t.Fatal("aceptó contenido que no corresponde al RootCID") }
	if _, err := almacen.Obtener(root); err != ErrContenidoNoDisponible { t.Fatalf("se esperaba contenido no disponible, err=%v", err) }

	if err := almacen.Guardar(root, contenido); err != nil { t.Fatal(err) }
	digest := root[len("rootcid:"):]
	path := filepath.Join(dir, digest[:2], digest)
	if err := os.WriteFile(path, []byte("corrupcion"), 0600); err != nil { t.Fatal(err) }
	if _, err := almacen.Obtener(root); err == nil { t.Fatal("devolvió un bloque corrupto") }
	if err := almacen.Guardar(root, contenido); err == nil { t.Fatal("sobrescribió silenciosamente un bloque corrupto") }
}

func TestAlmacenContenidoNoPermiteSobrescribirCID(t *testing.T) {
	almacen, err := NuevoAlmacenContenido("")
	if err != nil { t.Fatal(err) }
	contenido := []byte("inmutable")
	root := rootcid.CrearContenido(contenido)
	if err := almacen.Guardar(root, contenido); err != nil { t.Fatal(err) }
	if err := almacen.Guardar(root, []byte("otro contenido")); err == nil { t.Fatal("permitió sobrescritura") }
}
