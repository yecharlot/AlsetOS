package declarativo

import (
	"strings"
	"testing"
)

const sampleAOS = `
; Mostrador — todo es organismo, incluida la UI
(flow cafeteria-mostrador
  (version "1.0.0")
  (desc "Stock + venta + UI como ORGES")

  (orge stock
    (kind ORGES)
    (cap "stock.consultar" "stock.actualizar")
    (provides "stock.consulta" "1.0")
    (state (stock.unidades 10) (stock.umbral 3))
    (rule alerta_bajo
      (when (lte stock.unidades 3))
      (then stock.alerta true)))

  (orge venta
    (kind ORGES)
    (cap "venta.registrar")
    (requires "stock.consulta" "1.0")
    (state (venta.habilitada false))
    (rule listo
      (when (gt stock.unidades 0))
      (then venta.habilitada true)
      (indeterminate false)))

  (orge ui-mostrador
    (kind UI)
    (cap "ui.render")
    (ui
      (column
        (label "Cafetería")
        (badge (bind stock.unidades))
        (row
          (button "+" (emit stock.inc))
          (button "Vender" (emit venta.registrar))))))

  (assemble stock venta ui-mostrador)
  (surface ui-mostrador)
)
`

func TestParsearAOS_Mostrador(t *testing.T) {
	flow, err := ParsearAOS(sampleAOS)
	if err != nil {
		t.Fatal(err)
	}
	if flow.Name != "cafeteria-mostrador" {
		t.Fatalf("name=%s", flow.Name)
	}
	if len(flow.Manifiesto.Organismos) != 3 {
		t.Fatalf("orgs=%d", len(flow.Manifiesto.Organismos))
	}
	if flow.SurfaceOID == "" || !strings.Contains(flow.SurfaceOID, "ui-mostrador") {
		t.Fatalf("surface=%s", flow.SurfaceOID)
	}
	ui, ok := flow.Surfaces[flow.SurfaceOID]
	if !ok || ui.Type != "column" {
		t.Fatalf("ui=%+v", ui)
	}
	rt, err := NewRuntime("")
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.AplicarFlow(flow); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.EvalOrgRules("alset:org:stock:1"); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.EvalOrgRules("alset:org:venta:1"); err != nil {
		t.Fatal(err)
	}
}

func TestParsearAOS_SintaxisInvalida(t *testing.T) {
	_, err := ParsearAOS(`(flow solo`)
	if err == nil {
		t.Fatal("esperaba error")
	}
}
