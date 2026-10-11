package pedidos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yecharlot/AlsetOS/declarativo"
)

func TestResolvePedidoA(t *testing.T) {
	asm, err := declarativo.Resolve(AssemblyWithInventarioA())
	if err != nil {
		t.Fatal(err)
	}
	prov, ok := asm.ProviderFor["alset:org:pedido:1|inventario.consulta"]
	if !ok || prov != "alset:org:inventario:a" {
		t.Fatalf("provider inventario: got %s ok=%v", prov, ok)
	}
}

func TestAmbiguousRejected(t *testing.T) {
	_, err := declarativo.Resolve(AmbiguousAssembly())
	if err == nil {
		t.Fatal("esperaba rechazo por ambigüedad")
	}
}

func TestVerticalSlicePedido(t *testing.T) {
	dir := t.TempDir()
	rt, err := declarativo.NewRuntime(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.LoadAssembly(AssemblyWithInventarioA()); err != nil {
		t.Fatal(err)
	}

	// Pulse firmado
	p, err := rt.EmitPulse("pedido.registrar", "alset:org:pedido:1", "alset:org:auditoria:1", "qty=5", "nonce-1")
	if err != nil {
		t.Fatal(err)
	}
	if !declarativo.VerifyPulse(p) {
		t.Fatal("firma Ed25519 no verifica")
	}
	if _, err := rt.EmitPulse("pedido.registrar", "alset:org:pedido:1", "", "x", "nonce-1"); err == nil {
		t.Fatal("replay nonce debió fallar")
	}

	recs, err := rt.EvalOrgRules("alset:org:pedido:1")
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) < 3 {
		t.Fatalf("reglas: %d", len(recs))
	}
	viable, ok := rt.GetState("pedido.conclusion.viable")
	if !ok || viable != true {
		t.Fatalf("viable=%v ok=%v history=%v", viable, ok, rt.State.HistoryCopy())
	}

	// Aprobar requiere capacidad en organismo aprobación, no en pedido
	if err := rt.RequireCap("alset:org:pedido:1", "pedido.aprobar"); err == nil {
		t.Fatal("pedido no debe tener pedido.aprobar")
	}
	if err := rt.RequireCap("alset:org:aprobacion:1", "pedido.aprobar"); err != nil {
		t.Fatal(err)
	}
	rt.SetState("aprobacion.estado", "aprobado")
	if rt.AuditLen() == 0 {
		t.Fatal("audit vacío")
	}

	// Cambio de stock → solo conclusiones dirty se recalculan
	rt.SetState("inventario.stock", float64(2))
	dirtyRecs := rt.EvalDirty()
	if len(dirtyRecs) == 0 {
		t.Fatal("esperaba recálculo granular")
	}
	stockConc, _ := rt.GetState("pedido.conclusion.stock")
	if stockConc == "ok" {
		t.Fatalf("stock bajo no debió quedar ok: %v", stockConc)
	}

	if err := rt.Persist(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "runtime_state.json")); err != nil {
		t.Fatal(err)
	}

	rt2, _ := declarativo.NewRuntime(dir)
	if err := rt2.LoadAssembly(AssemblyWithInventarioA()); err != nil {
		t.Fatal(err)
	}
	if err := rt2.Restore(); err != nil {
		t.Fatal(err)
	}
	stock, ok := rt2.GetState("inventario.stock")
	if !ok {
		t.Fatal("stock no restaurado")
	}
	if stock.(float64) != 2 {
		t.Fatalf("stock restaurado=%v", stock)
	}
}

func TestSustituirProveedorInventario(t *testing.T) {
	// Misma declaración Pedido(); solo cambia el ORGES de inventario en el ensamblaje.
	rt, err := declarativo.NewRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.LoadAssembly(AssemblyWithInventarioB()); err != nil {
		t.Fatal(err)
	}
	prov, ok := rt.ProviderOID("alset:org:pedido:1", "inventario.consulta")
	if !ok || prov != "alset:org:inventario:b" {
		t.Fatalf("provider B: %s", prov)
	}
	// Pedido rules unchanged — stock B inicial es 3 < 5 → stock no ok
	_, _ = rt.EvalOrgRules("alset:org:pedido:1")
	stockConc, _ := rt.GetState("pedido.conclusion.stock")
	if stockConc == "ok" {
		t.Fatal("con inventario B stock=3 no debería ser ok para cantidad 5")
	}
	// Código del organismo Pedido (función Pedido()) no se modificó entre A y B.
}

func TestMindDoesNotGrantCap(t *testing.T) {
	rt, _ := declarativo.NewRuntime("")
	_ = rt.LoadAssembly(AssemblyWithInventarioA())
	// Inferencia viable=true no otorga pedido.aprobar al pedido
	_, _ = rt.EvalOrgRules("alset:org:pedido:1")
	if rt.HasCap("alset:org:pedido:1", "pedido.aprobar") {
		t.Fatal("inferencia no debe conceder capacidades")
	}
}

