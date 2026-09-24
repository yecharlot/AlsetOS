package pruebas

import (
	"testing"

	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/nodo"
)

func TestNodoEjecutaYRegistraOrganismo(t *testing.T) {
	n, err := nodo.Nuevo("nodo-prueba", t.TempDir()+"/registro.json")
	if err != nil { t.Fatal(err) }

	doc := manifiesto.Manifiesto{
		Nombre: "organismo-nodo",
		Zyrion: manifiesto.ConfiguracionZyrion{Estado: "si"},
		Capacidades: []string{"gene.ejecutar"},
		Genes: []string{"gene-saludo"},
	}

	resultado, err := n.Ejecutar(doc)
	if err != nil { t.Fatal(err) }
	if resultado.Organismo.RootCID == "" { t.Fatal("RootCID vacío") }

	entrada, existe := n.Registro.Obtener(resultado.Organismo.RootCID)
	if !existe { t.Fatal("organismo no registrado") }
	if entrada.Nombre != "organismo-nodo" {
		t.Fatalf("nombre inesperado: %s", entrada.Nombre)
	}
}
