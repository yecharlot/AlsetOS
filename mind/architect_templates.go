package mind

import (
	"fmt"
	"strings"
)

// processKind clasifica un nombre de proceso para plantilla ORGES específica.
type processKind string

const (
	kindInventario processKind = "inventario"
	kindAlmacen    processKind = "almacen"
	kindRecepcion  processKind = "recepcion"
	kindVentas     processKind = "ventas"
	kindCaja       processKind = "caja"
	kindIngresos   processKind = "ingresos"
	kindGastos     processKind = "gastos"
	kindAuditoria  processKind = "auditoria"
	kindUtilidades processKind = "utilidades"
	kindClientes   processKind = "clientes"
	kindPedidos    processKind = "pedidos"
	kindNomina     processKind = "nomina"
	kindGeneric    processKind = "generic"
)

func classifyProcess(name string) processKind {
	p := strings.ToLower(strings.TrimSpace(name))
	// Solo el nombre del proceso (no el texto global de la entrevista).
	switch {
	case containsAny(p, "recepcion", "recepción", "entrada de", "compra"):
		return kindRecepcion
	case containsAny(p, "almacen", "almacén", "bodega", "warehouse"):
		return kindAlmacen
	case containsAny(p, "inventario", "stock", "existencia"):
		return kindInventario
	case containsAny(p, "venta", "pos", "mostrador", "factur"):
		return kindVentas
	case containsAny(p, "caja", "cobro", "pago"):
		return kindCaja
	case containsAny(p, "ingreso", "revenue", "recaud"):
		return kindIngresos
	case containsAny(p, "gasto", "egreso", "costo", "expense"):
		return kindGastos
	case containsAny(p, "auditor", "trazabil", "log", "bitacora", "bitácora"):
		return kindAuditoria
	case containsAny(p, "utilidad", "beneficio", "margen", "profit"):
		return kindUtilidades
	case containsAny(p, "cliente", "customer", "crm"):
		return kindClientes
	case containsAny(p, "pedido", "orden", "order", "ruta", "delivery"):
		return kindPedidos
	case containsAny(p, "nomina", "nómina", "empleado", "payroll"):
		return kindNomina
	default:
		return kindGeneric
	}
}

func containsAny(s string, keys ...string) bool {
	for _, k := range keys {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

// writeProcessOrge emite un orge con estado y reglas propios del tipo de proceso.
func writeProcessOrge(b *strings.Builder, oid string, kind processKind) {
	b.WriteString(fmt.Sprintf("  (orge %s\n", oid))
	b.WriteString("    (kind ORGES)\n")

	switch kind {
	case kindInventario:
		b.WriteString(fmt.Sprintf("    (cap \"%s.consultar\" \"%s.actualizar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.stock\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.unidades 10) (%s.minimo 3) (%s.alerta false) (%s.eventos 0))\n", oid, oid, oid, oid))
		b.WriteString(fmt.Sprintf("    (rule %s-stock-bajo (when (lt %s.unidades %s.minimo)) (then %s.alerta true))\n", oid, oid, oid, oid))

	case kindAlmacen:
		b.WriteString(fmt.Sprintf("    (cap \"%s.consultar\" \"%s.mover\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.ubicacion\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.slots 20) (%s.ocupados 0) (%s.capacidad_ok true))\n", oid, oid, oid))
		b.WriteString(fmt.Sprintf("    (rule %s-capacidad (when (gte %s.ocupados %s.slots)) (then %s.capacidad_ok false))\n", oid, oid, oid, oid))

	case kindRecepcion:
		b.WriteString(fmt.Sprintf("    (cap \"%s.registrar\" \"%s.consultar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (requires \"inventario.stock\" \"1.0\")\n"))
		b.WriteString(fmt.Sprintf("    (state (%s.pendientes 0) (%s.recibidas 0) (%s.ultima \"\"))\n", oid, oid, oid))

	case kindVentas:
		b.WriteString(fmt.Sprintf("    (cap \"%s.registrar\" \"%s.consultar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (requires \"inventario.stock\" \"1.0\")\n"))
		b.WriteString(fmt.Sprintf("    (state (%s.total 0) (%s.ticket 0) (%s.eventos 0))\n", oid, oid, oid))

	case kindCaja:
		b.WriteString(fmt.Sprintf("    (cap \"%s.abrir\" \"%s.cerrar\" \"%s.consultar\")\n", oid, oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.abierta false) (%s.saldo 0) (%s.movimientos 0))\n", oid, oid, oid))

	case kindIngresos:
		b.WriteString(fmt.Sprintf("    (cap \"%s.registrar\" \"%s.consultar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.monto 0) (%s.periodos 0) (%s.ultimo 0))\n", oid, oid, oid))

	case kindGastos:
		b.WriteString(fmt.Sprintf("    (cap \"%s.registrar\" \"%s.consultar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.monto 0) (%s.categorias 0) (%s.presupuesto 0))\n", oid, oid, oid))
		b.WriteString(fmt.Sprintf("    (rule %s-sobre-presupuesto (when (gt %s.monto %s.presupuesto)) (then %s.alerta true))\n", oid, oid, oid, oid))
		b.WriteString(fmt.Sprintf("    (state (%s.alerta false))\n", oid))

	case kindAuditoria:
		b.WriteString(fmt.Sprintf("    (cap \"%s.registrar\" \"%s.consultar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.eventos 0) (%s.ultimo_actor \"\") (%s.cadena_ok true))\n", oid, oid, oid))

	case kindUtilidades:
		b.WriteString(fmt.Sprintf("    (cap \"%s.calcular\" \"%s.consultar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (requires \"ingresos.entity\" \"1.0\")\n"))
		b.WriteString(fmt.Sprintf("    (requires \"gastos.entity\" \"1.0\")\n"))
		b.WriteString(fmt.Sprintf("    (state (%s.margen 0) (%s.periodo \"\") (%s.positivo true))\n", oid, oid, oid))

	case kindClientes:
		b.WriteString(fmt.Sprintf("    (cap \"%s.registrar\" \"%s.consultar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.activos 0) (%s.nuevos 0))\n", oid, oid))

	case kindPedidos:
		b.WriteString(fmt.Sprintf("    (cap \"%s.crear\" \"%s.actualizar\" \"%s.consultar\")\n", oid, oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.abiertos 0) (%s.entregados 0) (%s.cancelados 0))\n", oid, oid, oid))

	case kindNomina:
		b.WriteString(fmt.Sprintf("    (cap \"%s.calcular\" \"%s.consultar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.empleados 0) (%s.periodo_pago \"\") (%s.total 0))\n", oid, oid, oid))

	default:
		b.WriteString(fmt.Sprintf("    (cap \"%s.operar\" \"%s.consultar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.activo true) (%s.eventos 0))\n", oid, oid))
	}
	b.WriteString("  )\n\n")
}

// writeBinds emite binds entre orges relacionados según sus kinds.
func writeBinds(b *strings.Builder, oids []string, kinds map[string]processKind) {
	has := func(k processKind) string {
		for _, id := range oids {
			if kinds[id] == k {
				return id
			}
		}
		return ""
	}
	inv := has(kindInventario)
	if inv == "" {
		inv = has(kindAlmacen)
	}
	ventas := has(kindVentas)
	recep := has(kindRecepcion)
	ing := has(kindIngresos)
	gas := has(kindGastos)
	util := has(kindUtilidades)
	caja := has(kindCaja)

	if ventas != "" && inv != "" {
		b.WriteString(fmt.Sprintf("  (bind %s %s %s.stock 1.0)\n", ventas, inv, inv))
	}
	if recep != "" && inv != "" {
		b.WriteString(fmt.Sprintf("  (bind %s %s %s.stock 1.0)\n", recep, inv, inv))
	}
	if caja != "" && ventas != "" {
		b.WriteString(fmt.Sprintf("  (bind %s %s %s.entity 1.0)\n", caja, ventas, ventas))
	}
	if util != "" && ing != "" {
		b.WriteString(fmt.Sprintf("  (bind %s %s %s.entity 1.0)\n", util, ing, ing))
	}
	if util != "" && gas != "" {
		b.WriteString(fmt.Sprintf("  (bind %s %s %s.entity 1.0)\n", util, gas, gas))
	}
}
