package mind

import (
	"strings"
	"testing"
)

func TestClassifyProcess(t *testing.T) {
	cases := map[string]processKind{
		"ventas":                    kindVentas,
		"Inventario de productos":   kindInventario,
		"almacén principal":         kindAlmacen,
		"recepción de inventario":   kindRecepcion,
		"ingresos":                  kindIngresos,
		"gastos operativos":         kindGastos,
		"auditoria":                 kindAuditoria,
		"utilidades":                kindUtilidades,
		"caja":                      kindCaja,
		"pedidos en ruta":           kindPedidos,
		"clientes":                  kindClientes,
		"nómina":                    kindNomina,
		"otro-proceso-xyz":          kindGeneric,
	}
	for in, want := range cases {
		if got := classifyProcess(in); got != want {
			t.Errorf("classifyProcess(%q)=%q want %q", in, got, want)
		}
	}
}

func TestBuildInterviewAOSNoStockOnGastos(t *testing.T) {
	a := &Agent{}
	in := InterviewInput{
		Empresa:       "Delicias en ruta",
		Sector:        "comercio",
		Procesos:      "ventas, inventario, almacen, recepción de inventario, ingresos, gastos, auditoria, utilidades",
		Funcionales:   "control de stock, ventas diarias, offline first",
		NoFuncionales: "offline first, auditoria, multi-usuario",
		Sistemas:      "PDF",
		WorkMode:      "model",
	}
	procs := splitList(in.Procesos)
	draft := a.buildInterviewAOS("delicias-en-ruta", in, procs)

	if strings.Contains(draft, "gastos.stock") {
		t.Fatal("gastos no debe tener stock:\n", draft)
	}
	if strings.Contains(draft, "auditoria.stock") {
		t.Fatal("auditoria no debe tener stock")
	}
	if strings.Contains(draft, "ingresos.stock") {
		t.Fatal("ingresos no debe tener stock")
	}
	if !strings.Contains(draft, "inventario.unidades") {
		t.Fatal("inventario debe tener unidades:\n", draft)
	}
	if !strings.Contains(draft, "ventas.total") {
		t.Fatal("ventas debe tener total")
	}
	if !strings.Contains(draft, "gastos.monto") {
		t.Fatal("gastos debe tener monto")
	}
	if !strings.Contains(draft, "(bind ") {
		t.Fatal("debe emitir binds entre orges relacionados")
	}
	t.Log(draft)
}
