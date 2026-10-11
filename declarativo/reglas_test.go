package declarativo

import (
	"path/filepath"
	"testing"
)

func TestAplicarReglasCafeteria(t *testing.T) {
	ruta := filepath.Join("..", "ejemplos", "reglas", "cafeteria.json")
	m, err := CargarReglas(ruta)
	if err != nil {
		t.Fatal(err)
	}
	rt, err := NewRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.AplicarReglas(m); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.EvalOrgRules("alset:org:stock:1"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := GenerarCodigo(m, dir); err != nil {
		t.Fatal(err)
	}
}
