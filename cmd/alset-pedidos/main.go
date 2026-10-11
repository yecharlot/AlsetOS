// Demostración ejecutable: dominio pedidos con ORGES declarativos.
//
//	go run ./cmd/alset-pedidos
//	go run ./cmd/alset-pedidos -provider=b
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yecharlot/AlsetOS/declarativo"
	"github.com/yecharlot/AlsetOS/ejemplos/pedidos"
)

func main() {
	provider := flag.String("provider", "a", "inventario provider: a | b")
	dataDir := flag.String("data", filepath.Join(os.TempDir(), "alset-pedidos-demo"), "directorio de persistencia")
	flag.Parse()

	_ = os.MkdirAll(*dataDir, 0o700)
	rt, err := declarativo.NewRuntime(*dataDir)
	must(err)

	var decls []declarativo.OrgDecl
	switch *provider {
	case "b":
		decls = pedidos.AssemblyWithInventarioB()
		fmt.Println("=== Ensamblaje con Inventario B (Pedido sin cambios de código) ===")
	default:
		decls = pedidos.AssemblyWithInventarioA()
		fmt.Println("=== Ensamblaje con Inventario A ===")
	}

	must(rt.LoadAssembly(decls))
	prov, _ := rt.ProviderOID("alset:org:pedido:1", "inventario.consulta")
	fmt.Println("Provider inventario.consulta →", prov)
	for _, b := range rt.Assembly.Bindings {
		fmt.Printf("  bind %s --[%s@%s]--> %s (%s)\n", b.ConsumerOID, b.Contract, b.Version, b.ProviderOID, b.Reason)
	}

	p, err := rt.EmitPulse("pedido.registrar", "alset:org:pedido:1", "alset:org:auditoria:1", "cantidad=5 costo=50", "demo-nonce-1")
	must(err)
	fmt.Println("Pulse firmado OK:", declarativo.VerifyPulse(p))

	recs, err := rt.EvalOrgRules("alset:org:pedido:1")
	must(err)
	fmt.Println("\n--- Inferencias ARL/Zyrion ---")
	for _, r := range recs {
		fmt.Printf("  rule=%s zyrion=%s then=%v skipped=%v reason=%s\n", r.RuleID, r.Zyrion, r.ThenValue, r.Skipped, r.Reason)
		for _, e := range r.Evidence {
			fmt.Println("    ·", e)
		}
	}

	fmt.Println("\n--- Estado relevante ---")
	for _, k := range []string{"inventario.stock", "inventario.proveedor", "presupuesto.disponible", "pedido.conclusion.stock", "pedido.conclusion.presupuesto", "pedido.conclusion.viable"} {
		v, _ := rt.GetState(k)
		fmt.Printf("  %s = %v\n", k, v)
	}

	fmt.Println("\n--- Autorización (Mind no concede caps) ---")
	fmt.Println("  pedido tiene pedido.aprobar?", rt.HasCap("alset:org:pedido:1", "pedido.aprobar"))
	fmt.Println("  aprobacion tiene pedido.aprobar?", rt.HasCap("alset:org:aprobacion:1", "pedido.aprobar"))

	fmt.Println("\n--- Cambio stock → recálculo granular ---")
	rt.SetState("inventario.stock", float64(1))
	for _, r := range rt.EvalDirty() {
		fmt.Printf("  dirty rule=%s zyrion=%s then=%v\n", r.RuleID, r.Zyrion, r.ThenValue)
	}

	must(rt.Persist())
	fmt.Println("\nPersistido en", filepath.Join(*dataDir, "runtime_state.json"))
	fmt.Println("Auditoría entradas:", rt.AuditLen())

	snap, _ := json.MarshalIndent(rt.State.Snapshot(), "", "  ")
	fmt.Println("\nSnapshot átomos:")
	fmt.Println(string(snap))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
