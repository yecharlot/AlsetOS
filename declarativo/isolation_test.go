package declarativo

import (
	"path/filepath"
	"testing"
)

func TestDeployProductionIsolatedDirs(t *testing.T) {
	dir := t.TempDir()
	f1, err := ParsearAOS(`(flow app-alpha (orge a (state (x.v 1))) (orge ui (kind UI) (ui (label "A"))) (assemble a ui) (surface ui))`)
	if err != nil {
		t.Fatal(err)
	}
	f2, err := ParsearAOS(`(flow app-beta (orge b (state (y.v 2))) (orge ui (kind UI) (ui (label "B"))) (assemble b ui) (surface ui))`)
	if err != nil {
		t.Fatal(err)
	}
	d1, err := DeployProduction(f1, dir)
	if err != nil {
		t.Fatal(err)
	}
	d2, err := DeployProduction(f2, dir)
	if err != nil {
		t.Fatal(err)
	}
	if d1 == d2 {
		t.Fatal("same dir")
	}
	if filepath.Base(d1) != "app-alpha" || filepath.Base(d2) != "app-beta" {
		t.Fatalf("bases %s %s", d1, d2)
	}
}
