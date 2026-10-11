package declarativo

import "testing"

func TestHandleEmitIncDecVenta(t *testing.T) {
	src := `
(flow t
  (orge stock (state (stock.unidades 5)) (cap "s"))
  (orge venta (state (venta.ultima null)) (cap "v"))
  (orge ui (kind UI) (ui (button "x" (emit stock.inc))))
  (assemble stock venta ui)
  (surface ui)
)`
	flow, err := ParsearAOS(src)
	if err != nil {
		t.Fatal(err)
	}
	rt, err := NewRuntime("")
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.AplicarFlow(flow); err != nil {
		t.Fatal(err)
	}
	res, err := rt.HandleEmit("stock.inc")
	if err != nil {
		t.Fatal(err)
	}
	st := res["state"].(map[string]interface{})
	if fmtNum(st["stock.unidades"]) != 6 {
		t.Fatalf("inc got %v", st["stock.unidades"])
	}
	_, err = rt.HandleEmit("venta.registrar")
	if err != nil {
		t.Fatal(err)
	}
	st = rt.State.Snapshot()
	if fmtNum(st["stock.unidades"]) != 5 {
		t.Fatalf("after venta got %v", st["stock.unidades"])
	}
}

func fmtNum(v interface{}) float64 {
	return toFloat(v)
}
