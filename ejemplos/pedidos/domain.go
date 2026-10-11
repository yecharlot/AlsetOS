// Package pedidos: dominio de pedidos/existencias como ORGES declarativos.
// El organismo Pedido requiere contratos; no importa la implementación concreta
// del inventario (inv-a vs inv-b).
package pedidos

import "github.com/yecharlot/AlsetOS/declarativo"

// InventarioA proveedor inicial de stock.
func InventarioA() declarativo.OrgDecl {
	return declarativo.OrgDecl{
		OID: "alset:org:inventario:a", Name: "InventarioA", Version: "1.0.0",
		Kind: "ORGES", Description: "Inventario almacén A",
		Capabilities: []string{"inventario.consultar", "inventario.actualizar"},
		Goals:        []string{"disponibilidad"},
		Contracts: []declarativo.ContractRef{
			{Name: "inventario.consulta", Version: "1.0", Role: "provides"},
		},
		InitialState: map[string]interface{}{
			"inventario.stock":     float64(10),
			"inventario.proveedor": "A",
		},
	}
}

// InventarioB segundo proveedor compatible (misma interfaz de contrato).
func InventarioB() declarativo.OrgDecl {
	return declarativo.OrgDecl{
		OID: "alset:org:inventario:b", Name: "InventarioB", Version: "1.1.0",
		Kind: "ORGES", Description: "Inventario almacén B (compatible 1.x)",
		Capabilities: []string{"inventario.consultar", "inventario.actualizar"},
		Goals:        []string{"disponibilidad"},
		Contracts: []declarativo.ContractRef{
			{Name: "inventario.consulta", Version: "1.0", Role: "provides"},
		},
		InitialState: map[string]interface{}{
			"inventario.stock":     float64(3),
			"inventario.proveedor": "B",
		},
	}
}

func Presupuesto() declarativo.OrgDecl {
	return declarativo.OrgDecl{
		OID: "alset:org:presupuesto:1", Name: "Presupuesto", Version: "1.0.0",
		Kind: "ORGES",
		Capabilities: []string{"presupuesto.consultar"},
		Goals:        []string{"control_gasto"},
		Contracts: []declarativo.ContractRef{
			{Name: "presupuesto.consulta", Version: "1.0", Role: "provides"},
		},
		InitialState: map[string]interface{}{
			"presupuesto.disponible": float64(1000),
		},
	}
}

func Auditoria() declarativo.OrgDecl {
	return declarativo.OrgDecl{
		OID: "alset:org:auditoria:1", Name: "Auditoria", Version: "1.0.0",
		Kind: "ORGES",
		Capabilities: []string{"auditoria.registrar"},
		Goals:        []string{"trazabilidad"},
		Contracts: []declarativo.ContractRef{
			{Name: "auditoria.registro", Version: "1.0", Role: "provides"},
		},
	}
}

// Aprobacion solo puede marcar aprobación si tiene capacidad explícita.
func Aprobacion() declarativo.OrgDecl {
	return declarativo.OrgDecl{
		OID: "alset:org:aprobacion:1", Name: "Aprobacion", Version: "1.0.0",
		Kind: "ORGES",
		Capabilities: []string{"pedido.aprobar"},
		Goals:        []string{"autorizacion"},
		Contracts: []declarativo.ContractRef{
			{Name: "pedido.aprobacion", Version: "1.0", Role: "provides"},
		},
		InitialState: map[string]interface{}{
			"aprobacion.estado": "pendiente",
		},
	}
}

// Pedido consumidor: requiere inventario, presupuesto, aprobación, auditoría.
// Reglas ARL estructuradas — no se modifican al cambiar de InventarioA a B.
func Pedido() declarativo.OrgDecl {
	return declarativo.OrgDecl{
		OID: "alset:org:pedido:1", Name: "Pedido", Version: "1.0.0",
		Kind: "ORGES", Description: "Gestión de pedido de compra",
		Capabilities: []string{"pedido.registrar", "pedido.consultar"},
		Goals:        []string{"cumplir_pedido", "no_sobregiro"},
		Contracts: []declarativo.ContractRef{
			{Name: "inventario.consulta", Version: "1.0", Role: "requires"},
			{Name: "presupuesto.consulta", Version: "1.0", Role: "requires"},
			{Name: "pedido.aprobacion", Version: "1.0", Role: "requires"},
			{Name: "auditoria.registro", Version: "1.0", Role: "requires"},
		},
		InitialState: map[string]interface{}{
			"pedido.cantidad": float64(5),
			"pedido.costo":    float64(50),
			"pedido.estado":   "nuevo",
		},
		Rules: []declarativo.Rule{
			{
				ID: "stock_suficiente",
				When: []declarativo.Condition{
					{Path: "inventario.stock", Op: "gte", Value: float64(5)},
					{Path: "pedido.cantidad", Op: "exists"},
				},
				ThenPath:  "pedido.conclusion.stock",
				ThenValue: "ok",
				ThenOnIndeterminate: "revisar",
			},
			{
				ID: "presupuesto_suficiente",
				When: []declarativo.Condition{
					{Path: "presupuesto.disponible", Op: "gte", Value: float64(50)},
				},
				ThenPath:  "pedido.conclusion.presupuesto",
				ThenValue: "ok",
			},
			{
				ID: "puede_aprobar_auto",
				When: []declarativo.Condition{
					{Path: "pedido.conclusion.stock", Op: "eq", Value: "ok"},
					{Path: "pedido.conclusion.presupuesto", Op: "eq", Value: "ok"},
				},
				ThenPath:  "pedido.conclusion.viable",
				ThenValue: true,
			},
		},
	}
}

// AssemblyWithInventarioA ensamblaje por defecto.
func AssemblyWithInventarioA() []declarativo.OrgDecl {
	return []declarativo.OrgDecl{InventarioA(), Presupuesto(), Aprobacion(), Auditoria(), Pedido()}
}

// AssemblyWithInventarioB sustituye proveedor sin cambiar Pedido().
func AssemblyWithInventarioB() []declarativo.OrgDecl {
	return []declarativo.OrgDecl{InventarioB(), Presupuesto(), Aprobacion(), Auditoria(), Pedido()}
}

// AmbiguousAssembly dos providers del mismo contrato → debe fallar Resolve.
func AmbiguousAssembly() []declarativo.OrgDecl {
	return []declarativo.OrgDecl{InventarioA(), InventarioB(), Presupuesto(), Aprobacion(), Auditoria(), Pedido()}
}
