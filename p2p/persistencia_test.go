package p2p

import (
	"path/filepath"
	"testing"
)

func TestPersistenciaOrganismo(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "organismos.json")
	original := map[string][]byte{"rootcid:demo": []byte("manifiesto")}

	if err := guardarOrganismos(ruta, original); err != nil { t.Fatal(err) }
	recuperado, err := cargarOrganismos(ruta)
	if err != nil { t.Fatal(err) }

	if string(recuperado["rootcid:demo"]) != "manifiesto" {
		t.Fatalf("contenido persistido incorrecto: %q", recuperado["rootcid:demo"])
	}
}
