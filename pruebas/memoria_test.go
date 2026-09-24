package pruebas

import (
	"testing"
	"github.com/yecharlot/AlsetOS/memoria"
)

func TestMemoriaPersiste(t *testing.T) {
	ruta := t.TempDir() + "/memoria.json"
	original := map[string]string{"estado":"activo","contador":"1"}
	if err := memoria.Guardar(ruta, original); err != nil { t.Fatal(err) }
	recuperada, err := memoria.Cargar(ruta)
	if err != nil { t.Fatal(err) }
	if recuperada["estado"] != "activo" || recuperada["contador"] != "1" { t.Fatalf("memoria incorrecta: %#v", recuperada) }
}
