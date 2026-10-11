package declarativo

import (
	"fmt"
	"strings"
)

// HandleEmit aplica funcionalidad real a un evento emitido desde la UI-orge.
// Convención: "stock.inc" | "stock.dec" | "dominio.campo=valor" | "venta.registrar"
func (rt *Runtime) HandleEmit(emit string) (map[string]interface{}, error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.State == nil {
		return nil, fmt.Errorf("sin estado")
	}
	emit = strings.TrimSpace(emit)
	if emit == "" {
		return nil, fmt.Errorf("emit vacío")
	}
	changed := map[string]interface{}{}

	switch {
	case strings.HasSuffix(emit, ".inc"):
		path := strings.TrimSuffix(emit, ".inc")
		// also try path without trailing and with domain style stock.unidades
		p := path
		if !strings.Contains(p, ".") {
			p = path + ".unidades"
		}
		v, _ := rt.State.Get(p)
		n := toFloat(v) + 1
		rt.State.Set(p, n)
		changed[p] = n
	case strings.HasSuffix(emit, ".dec"):
		path := strings.TrimSuffix(emit, ".dec")
		p := path
		if !strings.Contains(p, ".") {
			p = path + ".unidades"
		}
		v, _ := rt.State.Get(p)
		n := toFloat(v) - 1
		rt.State.Set(p, n)
		changed[p] = n
	case strings.Contains(emit, "="):
		parts := strings.SplitN(emit, "=", 2)
		rt.State.Set(strings.TrimSpace(parts[0]), parts[1])
		changed[parts[0]] = parts[1]
	default:
		// marcar evento y timestamp lógico
		key := "event." + emit
		rt.State.Set(key, true)
		changed[key] = true
		// venta.registrar: decrementa stock.unidades si existe
		if emit == "venta.registrar" {
			if v, ok := rt.State.Get("stock.unidades"); ok {
				n := toFloat(v) - 1
				if n < 0 {
					n = 0
				}
				rt.State.Set("stock.unidades", n)
				changed["stock.unidades"] = n
				rt.State.Set("venta.ultima", emit)
				changed["venta.ultima"] = emit
			}
		}
	}
	rt.audit("ui.emit", "studio", emit, fmt.Sprint(changed))
	// re-eval rules all orgs
	if rt.Assembly != nil {
		for _, o := range rt.Assembly.Orgs {
			for _, rule := range o.Rules {
				_ = rt.State.EvalRule(o.OID, rule)
			}
		}
	}
	snap := rt.State.Snapshot()
	return map[string]interface{}{"changed": changed, "state": snap}, nil
}

func toFloat(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		var f float64
		_, _ = fmt.Sscanf(x, "%f", &f)
		return f
	default:
		return 0
	}
}
