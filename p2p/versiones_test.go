package p2p

import (
	"path/filepath"
	"testing"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestDescriptorVersionFirmaYDerivacionDeIdentidad(t *testing.T) {
	id, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	root := rootcid.CrearContenido([]byte("revision uno"))
	d, err := CrearDescriptorVersion(id, "organismo:inventario", 1, root, "")
	if err != nil { t.Fatal(err) }
	if err := VerificarDescriptor(d); err != nil { t.Fatal(err) }
	d.RootCID = rootcid.CrearContenido([]byte("alterado"))
	if err := VerificarDescriptor(d); err == nil { t.Fatal("aceptó descriptor modificado después de firmar") }
}

func TestIndiceVersionesDetectaHuecosYConflictosSinElegirGanador(t *testing.T) {
	idA, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	idB, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	r1 := rootcid.CrearContenido([]byte("v1"))
	r2 := rootcid.CrearContenido([]byte("v2"))
	r2Fork := rootcid.CrearContenido([]byte("v2 alternativo"))
	r3 := rootcid.CrearContenido([]byte("v3"))
	indice, err := NuevoIndiceVersiones(filepath.Join(t.TempDir(), "versiones.json"))
	if err != nil { t.Fatal(err) }

	d2, err := CrearDescriptorVersion(idA, "organismo:x", 2, r2, r1)
	if err != nil { t.Fatal(err) }
	estado, err := indice.Registrar(d2)
	if err != nil { t.Fatal(err) }
	if estado != LinajeIncompleto { t.Fatalf("se esperaba linaje incompleto, estado=%s", estado) }

	d1, err := CrearDescriptorVersion(idA, "organismo:x", 1, r1, "")
	if err != nil { t.Fatal(err) }
	if estado, err = indice.Registrar(d1); err != nil || estado != LinajeCoherente {
		t.Fatalf("al completar antecesor: estado=%s err=%v", estado, err)
	}

	fork, err := CrearDescriptorVersion(idB, "organismo:x", 2, r2Fork, r1)
	if err != nil { t.Fatal(err) }
	if estado, err = indice.Registrar(fork); err != nil || estado != LinajeConflicto {
		t.Fatalf("se esperaba conflicto explícito, estado=%s err=%v", estado, err)
	}
	if len(indice.Versiones("organismo:x")) != 3 { t.Fatal("el índice debe conservar ambas ramas y el antecesor") }

	d3, err := CrearDescriptorVersion(idA, "organismo:x", 3, r3, r2)
	if err != nil { t.Fatal(err) }
	if estado, err = indice.Registrar(d3); err != nil || estado != LinajeConflicto {
		t.Fatalf("una nueva revisión no debe ocultar el fork, estado=%s err=%v", estado, err)
	}

	reabierto, err := NuevoIndiceVersiones(indice.ruta)
	if err != nil { t.Fatal(err) }
	if estado := reabierto.Estado("organismo:x"); estado != LinajeConflicto {
		t.Fatalf("el conflicto debe persistir tras reinicio, estado=%s", estado)
	}
}

func TestIndiceVersionesRegistroIdempotente(t *testing.T) {
	id, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	root := rootcid.CrearContenido([]byte("mismo contenido"))
	d, err := CrearDescriptorVersion(id, "organismo:idempotente", 1, root, "")
	if err != nil { t.Fatal(err) }
	indice, err := NuevoIndiceVersiones("")
	if err != nil { t.Fatal(err) }
	if _, err := indice.Registrar(d); err != nil { t.Fatal(err) }
	if _, err := indice.Registrar(d); err != nil { t.Fatal(err) }
	if got := len(indice.Versiones("organismo:idempotente")); got != 1 { t.Fatalf("se esperaba una entrada, got=%d", got) }
}
