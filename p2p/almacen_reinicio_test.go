package p2p

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestNodoReiniciaConContenidoVerificado(t *testing.T) {
	id, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	dir := t.TempDir()
	rutaIndice := filepath.Join(dir, "organismos.json")
	contenido := []byte(`{"nombre":"persistente","version":"1"}`)
	root := rootcid.CrearContenido(contenido)

	almacen, err := NuevoAlmacenContenido(filepath.Join(dir, "bloques"))
	if err != nil { t.Fatal(err) }
	if err := almacen.Guardar(root, contenido); err != nil { t.Fatal(err) }
	if err := guardarOrganismos(rutaIndice, map[string][]byte{root: contenido}); err != nil { t.Fatal(err) }

	reiniciado, err := NuevoConEstado(id, "/ip4/127.0.0.1/tcp/0", rutaIndice)
	if err != nil { t.Fatal(err) }
	defer reiniciado.Cerrar()
	disponible, err := reiniciado.DisponibleLocal(root)
	if err != nil || !disponible { t.Fatalf("bloque no disponible tras reinicio: %v %v", disponible, err) }
	recuperado, err := reiniciado.RecuperarOrganismo(context.Background(), root)
	if err != nil { t.Fatal(err) }
	if string(recuperado) != string(contenido) { t.Fatalf("bloque recuperado distinto: %s", recuperado) }
}

func TestNodoRechazaBloquePersistidoCorruptoAlIniciar(t *testing.T) {
	id, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	dir := t.TempDir()
	rutaIndice := filepath.Join(dir, "organismos.json")
	contenido := []byte("definicion original")
	root := rootcid.CrearContenido(contenido)
	almacen, err := NuevoAlmacenContenido(filepath.Join(dir, "bloques"))
	if err != nil { t.Fatal(err) }
	if err := almacen.Guardar(root, contenido); err != nil { t.Fatal(err) }
	if err := guardarOrganismos(rutaIndice, map[string][]byte{root: contenido}); err != nil { t.Fatal(err) }
	digest := root[len("rootcid:"):]
	if err := os.WriteFile(filepath.Join(dir, "bloques", digest[:2], digest), []byte("definicion manipulada"), 0600); err != nil { t.Fatal(err) }
	if nodo, err := NuevoConEstado(id, "/ip4/127.0.0.1/tcp/0", rutaIndice); err == nil {
		_ = nodo.Cerrar()
		t.Fatal("el nodo inició aceptando un bloque persistido corrupto")
	}
}
