package declarativo

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/yecharlot/AlsetOS/lispai"
)

// UINode IR de interfaz declarativa (la UI también es organismo / orge).
// Compatible en espíritu con UiRuntime (panel, label, button, badge, …).
type UINode struct {
	Type  string            `json:"type"` // column|row|label|button|badge|panel|field|drawer|stack|tabs|splash|blob|image|qr|file|gradient|doclist
	Text  string            `json:"text,omitempty"`
	Bind  string            `json:"bind,omitempty"`
	Emit  string            `json:"emit,omitempty"`
	Src   string            `json:"src,omitempty"`   // image url / gradient / file accept
	Style string            `json:"style,omitempty"` // css hints: blur, animated, bottom, top
	Kids  []UINode          `json:"kids,omitempty"`
	Meta  map[string]string `json:"meta,omitempty"`
}


// TestCase aserción sobre estado tras ensamblar (ATS-001 verificación).
type TestCase struct {
	Name string
	Path string
	Op   string
	Want interface{}
}

// DeployTarget empaquetado multiplataforma desde el flow.
type DeployTarget struct {
	Name    string
	Devices []string
	OutDir  string
}

// ContractBind une consumer → provider por contrato (LEGO explícito).
type ContractBind struct {
	Consumer string `json:"consumer"`
	Provider string `json:"provider"`
	Contract string `json:"contract"`
	Version  string `json:"version"`
}

// FlowAOS resultado de compilar un programa LISPAI-AOS.
type FlowAOS struct {
	Name        string
	Version     string
	Description string
	Manifiesto  *ManifiestoReglas
	Surfaces    map[string]UINode // oid UI → árbol
	SurfaceOID  string            // superficie activa
	Assemble    []string          // oids a ensamblar
	Tests       []TestCase
	Deploy      *DeployTarget
	Binds       []ContractBind // uniones LEGO explícitas
}

// CargarAOS lee un archivo .aos / .lispai y lo compila a FlowAOS.
func CargarAOS(ruta string) (*FlowAOS, error) {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}
	return ParsearAOS(string(b))
}

// ParsearAOS compila fuente LISPAI-AOS.
func ParsearAOS(src string) (*FlowAOS, error) {
	tokens := lispai.TokenizarAOS(src)
	forms, err := lispai.ParsearAOS(tokens)
	if err != nil {
		return nil, err
	}
	flow := &FlowAOS{
		Version:    "1.0.0",
		Surfaces:   map[string]UINode{},
		Manifiesto: &ManifiestoReglas{Plataforma: "AlsetOS", Version: "1.0.0"},
	}
	for _, f := range forms {
		if err := compileTop(flow, f); err != nil {
			return nil, err
		}
	}
	if flow.Name == "" {
		return nil, fmt.Errorf("AOS: se requiere un (flow nombre …)")
	}
	flow.Manifiesto.App = flow.Name
	flow.Manifiesto.Version = flow.Version
	flow.Manifiesto.Descripcion = flow.Description
	if flow.Manifiesto.Meta == nil {
		flow.Manifiesto.Meta = map[string]string{}
	}
	flow.Manifiesto.Meta["source"] = "lispai-aos"
	if flow.SurfaceOID != "" {
		flow.Manifiesto.Meta["surface"] = flow.SurfaceOID
	}
	if len(flow.Manifiesto.Organismos) == 0 {
		return nil, fmt.Errorf("AOS: el flow no define ningún (orge …)")
	}
	return flow, nil
}

func compileTop(flow *FlowAOS, n lispai.AOSNode) error {
	list, ok := n.([]lispai.AOSNode)
	if !ok || len(list) == 0 {
		return nil
	}
	head := fmt.Sprint(list[0])
	switch head {
	case "flow":
		return compileFlow(flow, list)
	case "orge":
		return compileOrge(flow, list)
	default:
		return fmt.Errorf("AOS: form de top-level desconocida: %s (usa flow u orge)", head)
	}
}

func compileFlow(flow *FlowAOS, list []lispai.AOSNode) error {
	if len(list) < 2 {
		return fmt.Errorf("AOS: (flow nombre …)")
	}
	flow.Name = fmt.Sprint(list[1])
	for _, item := range list[2:] {
		l, ok := item.([]lispai.AOSNode)
		if !ok || len(l) == 0 {
			continue
		}
		switch fmt.Sprint(l[0]) {
		case "version":
			if len(l) > 1 {
				flow.Version = fmt.Sprint(l[1])
			}
		case "desc", "description":
			if len(l) > 1 {
				flow.Description = fmt.Sprint(l[1])
			}
		case "orge":
			if err := compileOrge(flow, l); err != nil {
				return err
			}
		case "assemble", "use":
			for _, x := range l[1:] {
				flow.Assemble = append(flow.Assemble, fmt.Sprint(x))
			}
		case "surface":
			if len(l) > 1 {
				flow.SurfaceOID = oidFor(fmt.Sprint(l[1]))
			}
		case "bind":
			// (bind consumer provider contract [version])
			if len(l) >= 4 {
				b := ContractBind{
					Consumer: oidFor(fmt.Sprint(l[1])),
					Provider: oidFor(fmt.Sprint(l[2])),
					Contract: fmt.Sprint(l[3]),
					Version:  "1.0",
				}
				if len(l) > 4 {
					b.Version = fmt.Sprint(l[4])
				}
				flow.Binds = append(flow.Binds, b)
			}
		case "test":
			// (test nombre (eq path value)|...)
			tc := TestCase{Name: "test"}
			if len(l) > 1 {
				tc.Name = fmt.Sprint(l[1])
			}
			for _, x := range l[2:] {
				xl, ok := x.([]lispai.AOSNode)
				if !ok || len(xl) < 3 {
					continue
				}
				tc.Op = fmt.Sprint(xl[0])
				tc.Path = fmt.Sprint(xl[1])
				tc.Want = parseValue(xl[2])
				flow.Tests = append(flow.Tests, tc)
			}
		case "deploy":
			// (deploy web (devices mobile tablet desktop) (out ./dist))
			d := &DeployTarget{Name: "web", Devices: []string{"mobile", "tablet", "desktop"}}
			if len(l) > 1 {
				d.Name = fmt.Sprint(l[1])
			}
			for _, x := range l[2:] {
				xl, ok := x.([]lispai.AOSNode)
				if !ok || len(xl) == 0 {
					continue
				}
				switch fmt.Sprint(xl[0]) {
				case "devices":
					d.Devices = nil
					for _, dv := range xl[1:] {
						d.Devices = append(d.Devices, fmt.Sprint(dv))
					}
				case "out":
					if len(xl) > 1 {
						d.OutDir = fmt.Sprint(xl[1])
					}
				}
			}
			flow.Deploy = d
		case "gene", "mind", "pulse":
			// registradas en meta para ATS-001 (extensibles por runtime)
			flow.ensureMeta()
			flow.Manifiesto.Meta[fmt.Sprint(l[0])] = fmt.Sprint(l[1:])
		default:
			return fmt.Errorf("AOS flow: directiva desconocida %s", l[0])
		}
	}
	return nil
}

func compileOrge(flow *FlowAOS, list []lispai.AOSNode) error {
	if len(list) < 2 {
		return fmt.Errorf("AOS: (orge nombre …)")
	}
	name := fmt.Sprint(list[1])
	oid := oidFor(name)
	org := OrgDecl{
		OID:          oid,
		Name:         name,
		Version:      "1.0.0",
		Kind:         "ORGES",
		InitialState: map[string]interface{}{},
	}
	for _, item := range list[2:] {
		l, ok := item.([]lispai.AOSNode)
		if !ok || len(l) == 0 {
			continue
		}
		switch fmt.Sprint(l[0]) {
		case "kind":
			if len(l) > 1 {
				org.Kind = strings.ToUpper(fmt.Sprint(l[1]))
			}
		case "desc", "description":
			if len(l) > 1 {
				org.Description = fmt.Sprint(l[1])
			}
		case "cap", "capabilities":
			for _, c := range l[1:] {
				org.Capabilities = append(org.Capabilities, fmt.Sprint(c))
			}
		case "goal", "goals":
			for _, g := range l[1:] {
				org.Goals = append(org.Goals, fmt.Sprint(g))
			}
		case "provides":
			// (provides name version)
			cr := ContractRef{Role: "provides"}
			if len(l) > 1 {
				cr.Name = fmt.Sprint(l[1])
			}
			if len(l) > 2 {
				cr.Version = fmt.Sprint(l[2])
			} else {
				cr.Version = "1.0"
			}
			org.Contracts = append(org.Contracts, cr)
		case "requires":
			cr := ContractRef{Role: "requires"}
			if len(l) > 1 {
				cr.Name = fmt.Sprint(l[1])
			}
			if len(l) > 2 {
				cr.Version = fmt.Sprint(l[2])
			} else {
				cr.Version = "1.0"
			}
			org.Contracts = append(org.Contracts, cr)
		case "state":
			for _, st := range l[1:] {
				sl, ok := st.([]lispai.AOSNode)
				if !ok || len(sl) < 2 {
					continue
				}
				org.InitialState[fmt.Sprint(sl[0])] = parseValue(sl[1])
			}
		case "rule":
			rule, err := compileRule(l)
			if err != nil {
				return err
			}
			org.Rules = append(org.Rules, rule)
		case "ui", "surface":
			node, err := compileUI(l[1:])
			if err != nil {
				return err
			}
			// UI tree may be single root or implicit column
			var root UINode
			if len(node) == 1 {
				root = node[0]
			} else {
				root = UINode{Type: "column", Kids: node}
			}
			flow.Surfaces[oid] = root
			org.Kind = "UI"
			if org.Meta == nil {
				org.Meta = map[string]string{}
			}
			org.Meta["ui"] = "true"
			if flow.SurfaceOID == "" {
				flow.SurfaceOID = oid
			}
		default:
			return fmt.Errorf("AOS orge %s: directiva desconocida %s", name, l[0])
		}
	}
	flow.Manifiesto.Organismos = append(flow.Manifiesto.Organismos, org)
	return nil
}

func compileRule(l []lispai.AOSNode) (Rule, error) {
	// (rule id (when (op path val)…) (then path val) [(indeterminate val)])
	r := Rule{}
	if len(l) < 2 {
		return r, fmt.Errorf("AOS: (rule id …)")
	}
	r.ID = fmt.Sprint(l[1])
	for _, item := range l[2:] {
		sl, ok := item.([]lispai.AOSNode)
		if !ok || len(sl) == 0 {
			continue
		}
		switch fmt.Sprint(sl[0]) {
		case "when":
			for _, w := range sl[1:] {
				cond, err := compileCond(w)
				if err != nil {
					return r, err
				}
				r.When = append(r.When, cond)
			}
		case "then":
			if len(sl) >= 3 {
				r.ThenPath = fmt.Sprint(sl[1])
				r.ThenValue = parseValue(sl[2])
			}
		case "indeterminate":
			if len(sl) >= 2 {
				r.ThenOnIndeterminate = parseValue(sl[1])
			}
		}
	}
	return r, nil
}

func compileCond(n lispai.AOSNode) (Condition, error) {
	// (op path value?)  op: eq ne gt gte lt lte exists missing
	l, ok := n.([]lispai.AOSNode)
	if !ok || len(l) < 2 {
		return Condition{}, fmt.Errorf("AOS: condición inválida")
	}
	c := Condition{Op: fmt.Sprint(l[0]), Path: fmt.Sprint(l[1])}
	if len(l) > 2 {
		c.Value = parseValue(l[2])
	}
	return c, nil
}

func compileUI(items []lispai.AOSNode) ([]UINode, error) {
	var out []UINode
	for _, it := range items {
		l, ok := it.([]lispai.AOSNode)
		if !ok || len(l) == 0 {
			continue
		}
		t := strings.ToLower(fmt.Sprint(l[0]))
		node := UINode{Type: t}
		switch t {
		case "column", "row", "panel", "drawer", "stack", "tabs", "top-tabs", "bottom-tabs", "splash", "blob", "gradient":
			// optional first string attrs: (drawer "left" ...) (splash "Bienvenido" ...)
			start := 1
			if len(l) > 1 {
				if _, isList := l[1].([]lispai.AOSNode); !isList {
					node.Text = fmt.Sprint(l[1])
					start = 2
				}
			}
			for _, extra := range l[start:] {
				el, ok := extra.([]lispai.AOSNode)
				if ok && len(el) >= 1 {
					h := strings.ToLower(fmt.Sprint(el[0]))
					if h == "style" && len(el) > 1 {
						node.Style = fmt.Sprint(el[1])
						continue
					}
					if h == "src" && len(el) > 1 {
						node.Src = fmt.Sprint(el[1])
						continue
					}
				}
			}
			// children = list nodes among remaining
			var childItems []lispai.AOSNode
			for _, extra := range l[start:] {
				if el, ok := extra.([]lispai.AOSNode); ok {
					h := strings.ToLower(fmt.Sprint(el[0]))
					if h == "style" || h == "src" || h == "emit" || h == "bind" {
						continue
					}
					childItems = append(childItems, extra)
				}
			}
			kids, err := compileUI(childItems)
			if err != nil {
				return nil, err
			}
			node.Kids = kids
		case "label", "button", "field", "tab":
			if len(l) > 1 {
				node.Text = fmt.Sprint(l[1])
			}
			for _, extra := range l[2:] {
				el, ok := extra.([]lispai.AOSNode)
				if !ok || len(el) < 2 {
					continue
				}
				switch fmt.Sprint(el[0]) {
				case "emit", "pulse":
					node.Emit = fmt.Sprint(el[1])
				case "bind":
					node.Bind = fmt.Sprint(el[1])
				case "style":
					node.Style = fmt.Sprint(el[1])
				case "src":
					node.Src = fmt.Sprint(el[1])
				}
			}
		case "badge", "qr":
			for _, extra := range l[1:] {
				el, ok := extra.([]lispai.AOSNode)
				if !ok || len(el) < 2 {
					if !ok {
						node.Text = fmt.Sprint(extra)
					}
					continue
				}
				switch fmt.Sprint(el[0]) {
				case "bind":
					node.Bind = fmt.Sprint(el[1])
				case "src":
					node.Src = fmt.Sprint(el[1])
				case "text":
					node.Text = fmt.Sprint(el[1])
				}
			}
		case "image":
			if len(l) > 1 {
				if el, ok := l[1].([]lispai.AOSNode); ok && len(el) >= 2 && fmt.Sprint(el[0]) == "src" {
					node.Src = fmt.Sprint(el[1])
				} else {
					node.Src = fmt.Sprint(l[1])
				}
			}
			for _, extra := range l[1:] {
				el, ok := extra.([]lispai.AOSNode)
				if !ok || len(el) < 2 {
					continue
				}
				if fmt.Sprint(el[0]) == "style" {
					node.Style = fmt.Sprint(el[1])
				}
				if fmt.Sprint(el[0]) == "src" {
					node.Src = fmt.Sprint(el[1])
				}
			}
		case "file", "file-picker":
			node.Type = "file"
			if len(l) > 1 {
				node.Text = fmt.Sprint(l[1])
			}
			node.Src = "image/*,.pdf,.txt"
			for _, extra := range l[1:] {
				el, ok := extra.([]lispai.AOSNode)
				if !ok || len(el) < 2 {
					continue
				}
				switch fmt.Sprint(el[0]) {
				case "accept":
					node.Src = fmt.Sprint(el[1])
				case "emit":
					node.Emit = fmt.Sprint(el[1])
				case "bind":
					node.Bind = fmt.Sprint(el[1])
				}
			}
		case "doclist":
			for _, extra := range l[1:] {
				el, ok := extra.([]lispai.AOSNode)
				if !ok || len(el) < 2 {
					continue
				}
				if fmt.Sprint(el[0]) == "bind" {
					node.Bind = fmt.Sprint(el[1])
				}
			}
		default:
			return nil, fmt.Errorf("AOS ui: tipo desconocido %s", t)
		}
		out = append(out, node)
	}
	return out, nil
}

func parseValue(n lispai.AOSNode) interface{} {
	s := fmt.Sprint(n)
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	if s == "null" || s == "nil" {
		return nil
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

func oidFor(name string) string {
	n := strings.TrimSpace(strings.ToLower(name))
	n = strings.ReplaceAll(n, " ", "-")
	if strings.HasPrefix(n, "alset:org:") {
		return n
	}
	return "alset:org:" + n + ":1"
}

// AplicarFlow ensambla el flow AOS en el runtime y registra superficies UI.
func (rt *Runtime) AplicarFlow(flow *FlowAOS) error {
	if flow == nil || flow.Manifiesto == nil {
		return fmt.Errorf("flow nulo")
	}
	if err := rt.AplicarReglas(flow.Manifiesto); err != nil {
		return err
	}
	rt.mu.Lock()
	defer rt.mu.Unlock()
	for oid, ui := range flow.Surfaces {
		rt.audit("ui.surface", oid, "registrada", ui.Type)
	}
	if flow.SurfaceOID != "" {
		rt.audit("ui.surface", flow.SurfaceOID, "activa", "lispai-aos")
	}
	return nil
}


func (flow *FlowAOS) ensureMeta() {
	if flow.Manifiesto == nil {
		return
	}
	if flow.Manifiesto.Meta == nil {
		flow.Manifiesto.Meta = map[string]string{}
	}
}
