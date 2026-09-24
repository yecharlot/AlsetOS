package autonomia

import (
	"path/filepath"
	"testing"
	"time"
)

func TestRegistroPersistente(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "placement.json")
	registro, err := NuevoRegistro(ruta)
	if err != nil {
		t.Fatal(err)
	}

	esperado := Colocacion{
		RootCID:      "rootcid:abc",
		Primario:     "node-a",
		Replicas:     []string{"node-b", "node-c"},
		Actualizado:  time.Now().UTC(),
	}
	if err := registro.Guardar(esperado); err != nil {
		t.Fatal(err)
	}

	recargado, err := NuevoRegistro(ruta)
	if err != nil {
		t.Fatal(err)
	}
	obtenido, ok := recargado.Obtener(esperado.RootCID)
	if !ok {
		t.Fatal("colocación no recuperada")
	}
	if obtenido.Primario != esperado.Primario {
		t.Fatalf("primario incorrecto: %s", obtenido.Primario)
	}
	if len(obtenido.Replicas) != 2 {
		t.Fatalf("réplicas incorrectas: %#v", obtenido.Replicas)
	}
}

func TestFiltrarReplica(t *testing.T) {
	resultado := filtrarReplica([]string{"a", "b", "a"}, "a")
	if len(resultado) != 1 || resultado[0] != "b" {
		t.Fatalf("resultado inesperado: %#v", resultado)
	}
}
