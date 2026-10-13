package mind

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Agent struct {
	Name string
}

func NewAgent() *Agent {
	return &Agent{Name: "AlsetInspector"}
}

type AnalyzeRequest struct {
	Source  string `json:"source"`
	Message string `json:"message"`
	Cursor  int    `json:"cursor"`
	Mode    string `json:"mode"` // analyze | suggest | chat | complete | build
}

type Suggestion struct {
	Title   string `json:"title"`
	Kind    string `json:"kind"`
	Snippet string `json:"snippet"`
	Reason  string `json:"reason"`
}

type AnalyzeResponse struct {
	Ok          bool         `json:"ok"`
	Reply       string       `json:"reply"`
	Suggestions []Suggestion `json:"suggestions"`
	Diagnostics []string     `json:"diagnostics"`
	Completions []string     `json:"completions,omitempty"`
	SurfaceHints []string    `json:"surface_hints,omitempty"`
	BuiltAOS    string       `json:"built_aos,omitempty"`
	Agent       string       `json:"agent"`
}

var (
	reOrge     = regexp.MustCompile(`\(orge\s+([a-zA-Z0-9_-]+)`)
	reRule     = regexp.MustCompile(`\(rule\s+([a-zA-Z0-9_-]+)`)
	reProvides = regexp.MustCompile(`\(provides\s+"([^"]+)"`)
	reRequires = regexp.MustCompile(`\(requires\s+"([^"]+)"`)
	reEmit     = regexp.MustCompile(`\(emit\s+([a-zA-Z0-9_.=-]+)`)
	reUI       = regexp.MustCompile(`\(kind\s+UI\)`)
	reFlow     = regexp.MustCompile(`\(flow\s+([a-zA-Z0-9_-]+)`)
	reTest     = regexp.MustCompile(`\(test\s+`)
	reDeploy   = regexp.MustCompile(`\(deploy\s+`)
	reAssemble = regexp.MustCompile(`\(assemble\s+`)
	reSurface  = regexp.MustCompile(`\(surface\s+([a-zA-Z0-9_-]+)`)
	reParen    = regexp.MustCompile(`[()]`)
)

func (a *Agent) Handle(req AnalyzeRequest) AnalyzeResponse {
	src := req.Source
	msg := strings.TrimSpace(req.Message)
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		if msg != "" {
			mode = "chat"
		} else {
			mode = "analyze"
		}
	}
	diags := a.diagnostics(src)
	sugs := a.suggestions(src, msg)
	resp := AnalyzeResponse{Ok: true, Agent: a.Name, Diagnostics: diags, Suggestions: sugs}

	// Domain build from natural language
	if mode == "build" || looksLikeDomainSpec(msg) {
		built := a.buildFromDomain(msg)
		if built != "" {
			resp.BuiltAOS = built
			resp.Suggestions = append([]Suggestion{{
				Title: "Flow generado desde dominio",
				Kind:  "build",
				Snippet: built,
				Reason: "Traducción lenguaje natural → LISPAI-AOS",
			}}, resp.Suggestions...)
			resp.Reply = "**AlsetInspector** generó un flow completo a partir de tu dominio.\nUsa «Aplicar en editor» para cargar el .aos.\n\n" + a.summarizeDomain(msg)
			return resp
		}
	}

	switch mode {
	case "complete":
		resp.Completions = a.completions(src)
		resp.Reply = "Completaciones contextuales ATS-001."
	case "chat":
		resp.Reply = a.chat(src, msg, sugs, diags)
	default:
		resp.Reply = a.analyzeReply(src, sugs, diags)
	}
	resp.SurfaceHints = a.surfaceHints(src)
	_ = os.Getenv // keep optional external hook site
	return resp
}

func looksLikeDomainSpec(msg string) bool {
	if len(msg) < 12 {
		return false
	}
	l := strings.ToLower(msg)
	keys := []string{"dominio", "necesito", "quiero", "casos de uso", "regla", "capacidad", "stock", "pedido", "venta", "cita", "caja", "tienda", "inventario", "build", "crea un", "genera"}
	for _, k := range keys {
		if strings.Contains(l, k) {
			return true
		}
	}
	return false
}

func (a *Agent) summarizeDomain(msg string) string {
	return "Entrada: " + strings.TrimSpace(msg)
}

func (a *Agent) buildFromDomain(msg string) string {
	l := strings.ToLower(msg)
	name := "app-dominio"
	switch {
	case strings.Contains(l, "pedido") || strings.Contains(l, "ruta") || strings.Contains(l, "delivery"):
		name = "pedidos-ruta"
		return flowPedidos(name)
	case strings.Contains(l, "cita") || strings.Contains(l, "medico") || strings.Contains(l, "médico") || strings.Contains(l, "agenda"):
		name = "citas-medicas"
		return flowCitas(name)
	case strings.Contains(l, "caja") || strings.Contains(l, "turno"):
		name = "caja-turnos"
		return flowCaja(name)
	case strings.Contains(l, "tienda") || strings.Contains(l, "catalog") || strings.Contains(l, "ecommerce"):
		name = "tienda-online"
		return flowTienda(name)
	case strings.Contains(l, "stock") || strings.Contains(l, "inventario") || strings.Contains(l, "venta") || strings.Contains(l, "cafeteria") || strings.Contains(l, "cafetería"):
		name = "cafeteria-mostrador"
		return flowStock(name)
	default:
		// generic from keywords
		return flowGeneric(name, msg)
	}
}

func flowStock(name string) string {
	return fmt.Sprintf(`(flow %s
  (version "1.0.0")
  (desc "Stock + venta generados por AlsetInspector")
  (orge stock
    (kind ORGES)
    (cap "stock.consultar" "stock.actualizar")
    (provides "stock.consulta" "1.0")
    (state (stock.unidades 10) (stock.umbral 3) (stock.alerta false))
    (rule alerta_bajo (when (lte stock.unidades 3)) (then stock.alerta true)))
  (orge venta
    (kind ORGES)
    (cap "venta.registrar")
    (requires "stock.consulta" "1.0")
    (state (venta.habilitada false))
    (rule listo (when (gt stock.unidades 0)) (then venta.habilitada true) (indeterminate false)))
  (orge ui-main
    (kind UI)
    (ui (column
      (splash "Operacion")
      (gradient)
      (badge (bind stock.unidades))
      (row
        (button "+" (emit stock.inc))
        (button "-" (emit stock.dec))
        (button "Vender" (emit venta.registrar))))))
  (assemble stock venta ui-main)
  (bind venta stock stock.consulta 1.0)
  (surface ui-main)
  (test stock-ini (eq stock.unidades 10))
  (deploy web (devices mobile tablet desktop)))
`, name)
}

func flowPedidos(name string) string {
	return fmt.Sprintf(`(flow %s
  (version "1.0.0")
  (desc "Pedidos en ruta — AlsetInspector")
  (orge pedidos
    (kind ORGES)
    (state (pedido.estado "nuevo") (pedido.total 0))
    (rule puede-ruta (when (eq pedido.estado "confirmado")) (then pedido.listo_ruta true)))
  (orge ui-main
    (kind UI)
    (ui (column
      (splash "Pedidos")
      (gradient)
      (badge (bind pedido.estado))
      (button "Confirmar" (emit pedido.estado=confirmado))
      (button "En ruta" (emit pedido.estado=en_ruta))
      (button "Entregado" (emit pedido.estado=entregado)))))
  (assemble pedidos ui-main)
  (surface ui-main)
  (test ini (eq pedido.estado "nuevo"))
  (deploy web (devices mobile tablet desktop)))
`, name)
}

func flowCitas(name string) string {
	return fmt.Sprintf(`(flow %s
  (version "1.0.0")
  (desc "Agenda de citas — AlsetInspector")
  (orge agenda
    (kind ORGES)
    (state (cita.cupos 5) (cita.reservadas 0) (cita.abierta true))
    (rule cerrar (when (gte cita.reservadas 5)) (then cita.abierta false)))
  (orge ui-main
    (kind UI)
    (ui (column
      (splash "Consultas")
      (badge (bind cita.reservadas))
      (button "Reservar" (emit cita.reservar)))))
  (assemble agenda ui-main)
  (surface ui-main)
  (test cupos (eq cita.cupos 5))
  (deploy web (devices mobile tablet desktop)))
`, name)
}

func flowCaja(name string) string {
	return fmt.Sprintf(`(flow %s
  (version "1.0.0")
  (desc "Caja y turnos — AlsetInspector")
  (orge caja
    (kind ORGES)
    (state (caja.abierta false) (caja.monto 0))
    (rule cobrar-ok (when (eq caja.abierta true)) (then caja.puede_cobrar true)))
  (orge ui-main
    (kind UI)
    (ui (column
      (splash "Caja")
      (badge (bind caja.monto))
      (button "Abrir" (emit caja.abierta=true))
      (button "Cerrar" (emit caja.abierta=false)))))
  (assemble caja ui-main)
  (surface ui-main)
  (test cerrada (eq caja.abierta false))
  (deploy web (devices mobile tablet desktop)))
`, name)
}

func flowTienda(name string) string {
	return fmt.Sprintf(`(flow %s
  (version "1.0.0")
  (desc "Tienda online — AlsetInspector")
  (orge catalogo
    (kind ORGES)
    (state (catalogo.items "A|B|C") (cart.count 0)))
  (orge ui-main
    (kind UI)
    (ui (column
      (splash "Tienda")
      (gradient)
      (badge (bind cart.count))
      (button "Agregar" (emit cart.inc)))))
  (assemble catalogo ui-main)
  (surface ui-main)
  (test cart (eq cart.count 0))
  (deploy web (devices mobile tablet desktop)))
`, name)
}

func flowGeneric(name, msg string) string {
	desc := strings.ReplaceAll(msg, `"`, "'")
	if len(desc) > 80 {
		desc = desc[:80]
	}
	return fmt.Sprintf(`(flow %s
  (version "1.0.0")
  (desc %q)
  (orge nucleo
    (kind ORGES)
    (cap "dominio.consultar")
    (state (dominio.listo false) (dominio.contador 0))
    (rule activar (when (eq dominio.listo false)) (then dominio.puede_activar true)))
  (orge ui-main
    (kind UI)
    (ui (column
      (splash "Dominio")
      (gradient)
      (badge (bind dominio.contador))
      (button "Activar" (emit dominio.listo=true))
      (button "+" (emit dominio.inc)))))
  (assemble nucleo ui-main)
  (surface ui-main)
  (test ini (eq dominio.listo false))
  (deploy web (devices mobile tablet desktop)))
`, name, desc)
}

func (a *Agent) diagnostics(src string) []string {
	var d []string
	if strings.TrimSpace(src) == "" {
		return []string{"Fuente vacía: empieza con (flow nombre …) o describe el dominio al Inspector."}
	}
	// balanced parens
	bal := 0
	for _, c := range src {
		if c == '(' {
			bal++
		} else if c == ')' {
			bal--
		}
		if bal < 0 {
			d = append(d, "Paréntesis de más: ')' sin '(' emparejado — revisa sintaxis LISPAI.")
			break
		}
	}
	if bal > 0 {
		d = append(d, fmt.Sprintf("Faltan %d paréntesis de cierre ')' — sintaxis incompleta.", bal))
	}
	if !reFlow.MatchString(src) {
		d = append(d, "Falta (flow nombre).")
	}
	orges := reOrge.FindAllStringSubmatch(src, -1)
	if len(orges) == 0 {
		d = append(d, "No hay (orge …).")
	}
	hasUI := reUI.MatchString(src)
	if hasUI && !reSurface.MatchString(src) {
		d = append(d, "UI sin (surface oid) — el preview no activará superficie.")
	}
	if !reAssemble.MatchString(src) && len(orges) > 1 {
		d = append(d, "Varias orges sin (assemble …).")
	}
	provides := map[string]bool{}
	for _, m := range reProvides.FindAllStringSubmatch(src, -1) {
		provides[m[1]] = true
	}
	for _, m := range reRequires.FindAllStringSubmatch(src, -1) {
		if !provides[m[1]] {
			d = append(d, fmt.Sprintf("requires %q sin provides — añade proveedor o (bind).", m[1]))
		}
	}
	if !reTest.MatchString(src) {
		d = append(d, "Sin (test …).")
	}
	if !reDeploy.MatchString(src) {
		d = append(d, "Sin (deploy …).")
	}
	if hasUI && !strings.Contains(src, "(emit ") {
		d = append(d, "UI sin (emit …) — botones no mutarán estado.")
	}
	return d
}

func (a *Agent) suggestions(src, msg string) []Suggestion {
	var out []Suggestion
	if !reFlow.MatchString(src) {
		out = append(out, Suggestion{Title: "Scaffold mínimo", Kind: "orge", Reason: "Sin flow", Snippet: flowGeneric("mi-app", "scaffold")})
	}
	if !reUI.MatchString(src) {
		out = append(out, Suggestion{Title: "UI animada", Kind: "ui", Reason: "Sin superficie UI", Snippet: `  (orge ui-main (kind UI) (ui (column (splash "Hola") (gradient) (button "OK" (emit x.ok)))))
  (surface ui-main)
`})
	}
	if !reRule.MatchString(src) && reOrge.MatchString(src) {
		out = append(out, Suggestion{Title: "Regla umbral", Kind: "rule", Reason: "Sin ARL", Snippet: `    (rule alerta (when (lte stock.unidades 3)) (then stock.alerta true))
`})
	}
	if !reTest.MatchString(src) {
		out = append(out, Suggestion{Title: "Test inicial", Kind: "test", Reason: "Sin tests", Snippet: `  (test ini (eq dominio.listo false))
`})
	}
	if !reDeploy.MatchString(src) {
		out = append(out, Suggestion{Title: "Deploy", Kind: "deploy", Reason: "Sin deploy", Snippet: `  (deploy web (devices mobile tablet desktop))
`})
	}
	if len(out) > 6 {
		out = out[:6]
	}
	_ = msg
	return out
}

func (a *Agent) completions(src string) []string {
	if strings.Contains(src, "(when") && !strings.Contains(src, "(then") {
		return []string{`(then campo valor)`, `(then flag true)`}
	}
	return []string{
		`(rule nombre (when (eq campo valor)) (then campo2 valor2))`,
		`(state (campo 0) (flag false))`,
		`(button "Accion" (emit campo.accion))`,
		`(badge (bind campo.valor))`,
		`(splash "Titulo")`,
		`(assemble orge-a ui-main)`,
		`(surface ui-main)`,
		`(deploy web (devices mobile tablet desktop))`,
	}
}

func (a *Agent) surfaceHints(src string) []string {
	var h []string
	if reUI.MatchString(src) {
		h = append(h, "Preview: Aplicar flow o Live ON para ver UI.")
	}
	if reEmit.MatchString(src) {
		h = append(h, "Emits: botones mutan estado en preview.")
	}
	return h
}

func (a *Agent) analyzeReply(src string, sugs []Suggestion, diags []string) string {
	orges := reOrge.FindAllStringSubmatch(src, -1)
	var names []string
	for _, m := range orges {
		names = append(names, m[1])
	}
	var b strings.Builder
	b.WriteString("**AlsetInspector** · análisis continuo\n\n")
	if len(names) > 0 {
		b.WriteString("Orges: " + strings.Join(names, ", ") + "\n")
	} else {
		b.WriteString("Sin organismos aún.\n")
	}
	if len(diags) == 0 {
		b.WriteString("\n✓ Sin diagnósticos críticos.\n")
	} else {
		b.WriteString("\nDiagnósticos:\n")
		for _, d := range diags {
			b.WriteString("• " + d + "\n")
		}
	}
	b.WriteString(fmt.Sprintf("\nSugerencias: %d. Describe un dominio en el chat para generar el flow completo.\n", len(sugs)))
	return b.String()
}

func (a *Agent) chat(src, msg string, sugs []Suggestion, diags []string) string {
	if looksLikeDomainSpec(msg) {
		return a.Handle(AnalyzeRequest{Source: src, Message: msg, Mode: "build"}).Reply
	}
	l := strings.ToLower(msg)
	var b strings.Builder
	b.WriteString("**AlsetInspector**\n\n")
	switch {
	case strings.Contains(l, "syntax") || strings.Contains(l, "sintaxis") || strings.Contains(l, "error"):
		if len(diags) == 0 {
			b.WriteString("No veo errores de estructura. Revisa paréntesis y nombres de átomos.")
		} else {
			for _, d := range diags {
				b.WriteString("• " + d + "\n")
			}
		}
	case strings.Contains(l, "pulse") || strings.Contains(l, "pulso"):
		b.WriteString("Pulse: mensaje contractual orge→orge vía provides/requires + bind. Payload mínimo, versionado, idempotente.")
	default:
		b.WriteString(a.analyzeReply(src, sugs, diags))
		b.WriteString("\nDescribe el dominio (reglas, casos de uso) y genero el .aos completo.")
	}
	return b.String()
}
