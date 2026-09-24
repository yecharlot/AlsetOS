package pruebas

import (
	"os"
	"testing"

	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/zyrion"
)

func TestManifiestoEstadosZyrion(t *testing.T) {
	contenido := []byte(`{"nombre":"prueba","zyrion":{"estado":"incierto","estrategia":"evaluar"}}`)
	ruta := t.TempDir() + "/organismo.alset"
	if err := os.WriteFile(ruta, contenido, 0600); err != nil { t.Fatal(err) }

	documento, err := manifiesto.Cargar(ruta)
	if err != nil { t.Fatal(err) }

	estado, err := documento.EstadoZyrion()
	if err != nil { t.Fatal(err) }

	if estado != zyrion.Incierto {
		t.Fatalf("estado inesperado: %v", estado)
	}
}
