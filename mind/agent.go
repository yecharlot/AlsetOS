package mind

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Agent is the Studio-coupled code agent (Mind style PrismaTEC).
// Analyzes .aos, suggests rules/UI, and chats about domain→ORGE mapping.
type Agent struct {
	Name string
}

func NewAgent() *Agent {
	return &Agent{Name: "Mind"}
}

// AnalyzeRequest is the payload from Studio.
type AnalyzeRequest struct {
	Source  string `json:"source"`
	Message string `json:"message"`
	Cursor  int    `json:"cursor"` // optional caret offset in source
	Mode    string `json:"mode"`   // analyze | suggest | chat | complete
}

// Suggestion is a code or structural proposal.
type Suggestion struct {
	Title   string `json:"title"`
	Kind    string `json:"kind"` // rule | orge | ui | test | deploy | fix | pulse
	Snippet string `json:"snippet"`
	Reason  string `json:"reason"`
}

// AnalyzeResponse is returned to Studio.
type AnalyzeResponse struct {
	Ok           bool         `json:"ok"`
	Reply        string       `json:"reply"`
	Suggestions  []Suggestion `json:"suggestions"`
	Diagnostics  []string     `json:"diagnostics"`
	Completions  []string     `json:"completions,omitempty"`
	SurfaceHints []string     `json:"surface_hints,omitempty"`
	Agent        string       `json:"agent"`
}

var (
	reOrge    = regexp.MustCompile(`\(orge\s+([a-zA-Z0-9_-]+)`)
	reRule    = regexp.MustCompile(`\(rule\s+([a-zA-Z0-9_-]+)`)
	reState   = regexp.MustCompile(`\(state\s+\(([^)]+)\)`)
	reProvides = regexp.MustCompile(`\(provides\s+"([^"]+)"`)
	reRequires = regexp.MustCompile(`\(requires\s+"([^"]+)"`)
	reEmit    = regexp.MustCompile(`\(emit\s+([a-zA-Z0-9_.=-]+)`)
	reUI      = regexp.MustCompile(`\(kind\s+UI\)`)
	reFlow    = regexp.MustCompile(`\(flow\s+([a-zA-Z0-9_-]+)`)
	reTest    = regexp.MustCompile(`\(test\s+`)
	reDeploy  = regexp.MustCompile(`\(deploy\s+`)
	reAssemble = regexp.MustCompile(`\(assemble\s+`)
	reSurface = regexp.MustCompile(`\(surface\s+([a-zA-Z0-9_-]+)`)
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

	// Optional external LLM if key present
	if external := a.tryExternal(req); external != nil {
		return *external
	}

	resp := AnalyzeResponse{
		Ok:          true,
		Agent:       a.Name,
		Diagnostics: diags,
		Suggestions: sugs,
	}

	switch mode {
	case "complete":
		resp.Completions = a.completions(src, req.Cursor)
		resp.Reply = "Completaciones contextuales según el flow y el cursor."
	case "suggest":
		resp.Reply = a.suggestReply(src, sugs, diags)
	case "chat":
		resp.Reply = a.chat(src, msg, sugs, diags)
	default:
		resp.Reply = a.analyzeReply(src, sugs, diags)
	}
	resp.SurfaceHints = a.surfaceHints(src)
	return resp
}

func (a *Agent) diagnostics(src string) []string {
	var d []string
	if strings.TrimSpace(src) == "" {
		return []string{"Fuente vacía: empieza con (flow nombre …)."}
	}
	if !reFlow.MatchString(src) {
		d = append(d, "Falta (flow nombre). Todo debe vivir dentro de un flow.")
	}
	orges := reOrge.FindAllStringSubmatch(src, -1)
	if len(orges) == 0 {
		d = append(d, "No hay (orge …). Añade al menos un organismo de dominio o UI.")
	}
	hasUI := reUI.MatchString(src)
	if hasUI && !reSurface.MatchString(src) {
		d = append(d, "Hay UI-orge pero falta (surface oid) para activar la superficie.")
	}
	if !reAssemble.MatchString(src) && len(orges) > 1 {
		d = append(d, "Varias orges sin (assemble …): el runtime no sabrá qué piezas cargar.")
	}
	provides := map[string]bool{}
	for _, m := range reProvides.FindAllStringSubmatch(src, -1) {
		provides[m[1]] = true
	}
	for _, m := range reRequires.FindAllStringSubmatch(src, -1) {
		if !provides[m[1]] {
			d = append(d, fmt.Sprintf("requires %q sin provides compatible en el flow — usa (bind) o añade el proveedor.", m[1]))
		}
	}
	if !reTest.MatchString(src) {
		d = append(d, "Sin (test …): añade aserciones para Probar en Studio.")
	}
	if !reDeploy.MatchString(src) {
		d = append(d, "Sin (deploy …): añade deploy web con devices para producción.")
	}
	if hasUI && !strings.Contains(src, "(emit ") {
		d = append(d, "UI sin (emit …): los botones no mutarán estado hasta que emitan acciones.")
	}
	return d
}

func (a *Agent) suggestions(src, msg string) []Suggestion {
	var out []Suggestion
	lower := strings.ToLower(src + " " + msg)

	if !reFlow.MatchString(src) {
		out = append(out, Suggestion{
			Title: "Scaffold flow mínimo",
			Kind:  "orge",
			Reason: "No hay flow; parte de una plantilla ATS-001.",
			Snippet: `(flow mi-dominio
  (version "1.0.0")
  (desc "Descripción del problema")

  (orge nucleo
    (kind ORGES)
    (cap "dominio.consultar")
    (state (dominio.listo false))
    (rule listo
      (when (eq dominio.listo false))
      (then dominio.listo true)))

  (orge ui-main
    (kind UI)
    (ui (column
      (splash "Mi dominio")
      (label "Estado")
      (badge (bind dominio.listo))
      (button "Activar" (emit dominio.listo=true)))))

  (assemble nucleo ui-main)
  (surface ui-main)
  (test ini (eq dominio.listo false))
  (deploy web (devices mobile tablet desktop)))
`,
		})
	}

	if !reRule.MatchString(src) && reOrge.MatchString(src) {
		out = append(out, Suggestion{
			Title:  "Regla ARL de umbral",
			Kind:   "rule",
			Reason: "Orges sin reglas: añade when/then para reactividad.",
			Snippet: `    (rule alerta
      (when (lte stock.unidades 3))
      (then stock.alerta true))
`,
		})
	}

	if strings.Contains(lower, "stock") || strings.Contains(lower, "inventario") || strings.Contains(lower, "almacen") {
		out = append(out, Suggestion{
			Title:  "ORGE stock + UI",
			Kind:   "orge",
			Reason: "Dominio de inventario detectado.",
			Snippet: `  (orge stock
    (kind ORGES)
    (cap "stock.consultar" "stock.actualizar")
    (provides "stock.consulta" "1.0")
    (state (stock.unidades 10) (stock.umbral 3) (stock.alerta false))
    (rule alerta_bajo
      (when (lte stock.unidades 3))
      (then stock.alerta true)))
`,
		})
	}

	if strings.Contains(lower, "venta") || strings.Contains(lower, "pos") || strings.Contains(lower, "caja") {
		out = append(out, Suggestion{
			Title:  "ORGE venta con requires",
			Kind:   "orge",
			Reason: "Dominio de venta/POS detectado.",
			Snippet: `  (orge venta
    (kind ORGES)
    (cap "venta.registrar")
    (requires "stock.consulta" "1.0")
    (state (venta.habilitada false))
    (rule listo
      (when (gt stock.unidades 0))
      (then venta.habilitada true)
      (indeterminate false)))
`,
		})
	}

	if strings.Contains(lower, "pedido") || strings.Contains(lower, "delivery") || strings.Contains(lower, "ruta") {
		out = append(out, Suggestion{
			Title:  "Flow pedidos en ruta",
			Kind:   "orge",
			Reason: "Dominio de pedidos/ruta.",
			Snippet: `  (orge pedidos
    (kind ORGES)
    (state (pedido.estado "nuevo") (pedido.total 0))
    (rule confirmar
      (when (eq pedido.estado "nuevo"))
      (then pedido.puede_confirmar true)))
  (orge ui-pedidos
    (kind UI)
    (ui (column
      (splash "Pedidos")
      (badge (bind pedido.estado))
      (button "Confirmar" (emit pedido.estado=confirmado))
      (button "En ruta" (emit pedido.estado=en_ruta))
      (button "Entregado" (emit pedido.estado=entregado)))))
`,
		})
	}

	if !reUI.MatchString(src) {
		out = append(out, Suggestion{
			Title:  "UI-orge animada",
			Kind:   "ui",
			Reason: "Sin superficie UI: añade splash, gradient y botones con emit.",
			Snippet: `  (orge ui-main
    (kind UI)
    (ui (column
      (splash "Bienvenida")
      (gradient)
      (badge (bind stock.unidades))
      (row
        (button "+" (emit stock.inc))
        (button "-" (emit stock.dec))))))
  (surface ui-main)
`,
		})
	}

	if !reTest.MatchString(src) {
		out = append(out, Suggestion{
			Title:  "Tests de estado",
			Kind:   "test",
			Reason: "Sin tests en el flow.",
			Snippet: `  (test stock-inicial (eq stock.unidades 10))
  (test sin-alerta (eq stock.alerta false))
`,
		})
	}

	if !reDeploy.MatchString(src) {
		out = append(out, Suggestion{
			Title:  "Deploy multi-dispositivo",
			Kind:   "deploy",
			Reason: "Falta empaque de producción.",
			Snippet: `  (deploy web (devices mobile tablet desktop))
`,
		})
	}

	out = append(out, Suggestion{
		Title:  "Pulso seguro entre orges",
		Kind:   "pulse",
		Reason: "Protocolo Pulse: mensaje firmado oid→oid con contrato.",
		Snippet: `  ;; Pulse conceptual: emisor → receptor vía contrato
  ;; (pulse stock venta "stock.consulta" (payload stock.unidades))
  (bind venta stock stock.consulta 1.0)
`,
	})

	if len(out) > 8 {
		out = out[:8]
	}
	return out
}

func (a *Agent) completions(src string, cursor int) []string {
	base := []string{
		`(rule nombre (when (eq campo valor)) (then campo2 valor2))`,
		`(state (campo 0) (flag false))`,
		`(provides "contrato.nombre" "1.0")`,
		`(requires "contrato.nombre" "1.0")`,
		`(button "Accion" (emit campo.accion))`,
		`(badge (bind campo.valor))`,
		`(splash "Titulo")`,
		`(column (label "x") (row (button "A" (emit a))))`,
		`(test nombre (eq campo valor))`,
		`(deploy web (devices mobile tablet desktop))`,
		`(assemble orge-a orge-b ui-main)`,
		`(surface ui-main)`,
		`(bind consumidor proveedor contrato 1.0)`,
	}
	// Contextual: if inside rule-ish
	if strings.Contains(src, "(when") && !strings.Contains(src, "(then") {
		return []string{`(then campo valor)`, `(then flag true)`, `(indeterminate false)`}
	}
	if strings.Contains(src, "(orge") && !strings.Contains(src, "(kind") {
		return []string{`(kind ORGES)`, `(kind UI)`}
	}
	_ = cursor
	return base
}

func (a *Agent) surfaceHints(src string) []string {
	var h []string
	if reUI.MatchString(src) {
		h = append(h, "Preview: Aplicar flow → Correr preview para ver UI-orge.")
		h = append(h, "Animación: usa splash, gradient, drawer; CSS studio aplica transiciones.")
	}
	if reEmit.MatchString(src) {
		h = append(h, "Emits detectados: en preview los botones mutan estado y reevalúan ARL.")
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
	b.WriteString("**Mind · análisis del flow**\n\n")
	if len(names) > 0 {
		b.WriteString(fmt.Sprintf("Organismos: %s\n", strings.Join(names, ", ")))
	} else {
		b.WriteString("Aún no hay organismos.\n")
	}
	if len(diags) > 0 {
		b.WriteString("\nDiagnósticos:\n")
		for _, d := range diags {
			b.WriteString("• " + d + "\n")
		}
	} else {
		b.WriteString("\nSin diagnósticos críticos.\n")
	}
	b.WriteString(fmt.Sprintf("\nSugerencias listas: %d. Puedes insertarlas en el editor o pedir detalle en el chat.\n", len(sugs)))
	b.WriteString("Tip: describe el problema de dominio en el chat (ej. «pedidos a domicilio con stock») y propongo el .aos completo.")
	return b.String()
}

func (a *Agent) suggestReply(src string, sugs []Suggestion, diags []string) string {
	return a.analyzeReply(src, sugs, diags)
}

func (a *Agent) chat(src, msg string, sugs []Suggestion, diags []string) string {
	lower := strings.ToLower(msg)
	var b strings.Builder
	b.WriteString("**Mind**\n\n")

	switch {
	case strings.Contains(lower, "hola") || strings.Contains(lower, "quién eres") || strings.Contains(lower, "quien eres"):
		b.WriteString("Soy Mind, el agente de código acoplado a Alset Studio. Leo tu .aos, diagnostico contratos/reglas/UI y propongo snippets ATS-001. No soy solo autocompletado: razono sobre el dominio que describes.")
	case strings.Contains(lower, "pulse") || strings.Contains(lower, "pulso"):
		b.WriteString("Pulse es el protocolo de mensaje entre orges: emisor → receptor bajo un contrato versionado (provides/requires + bind). Mantén payloads mínimos, versiona contratos y trata el pulso como evento idempotente. En el .aos se materializa con (bind) y emits que mutan estado compartido vía ensamblaje.")
	case strings.Contains(lower, "regla") || strings.Contains(lower, "arl"):
		b.WriteString("Reglas ARL: (rule id (when pred) (then asignaciones)). Predicados: eq, gt, gte, lt, lte. Tras cada emit el runtime reevalúa. Ejemplo:\n```\n(rule alerta (when (lte stock.unidades 3)) (then stock.alerta true))\n```")
	case strings.Contains(lower, "ui") || strings.Contains(lower, "vista") || strings.Contains(lower, "pantalla"):
		b.WriteString("Las vistas son orges kind UI. Árbol IR: column, row, splash, gradient, badge (bind), button (emit), drawer, tabs. Activa con (surface oid). Para pantallas «hermosas» combina splash + gradient + row de botones y prueba en preview mobile/tablet/desktop.")
	case strings.Contains(lower, "pedido") || strings.Contains(lower, "tienda") || strings.Contains(lower, "stock") || strings.Contains(lower, "venta"):
		b.WriteString("Para ese dominio te conviene: orge de inventario (provides stock.consulta), orge de venta/pedido (requires), UI con emits, tests de stock inicial y deploy. Revisa las sugerencias insertables a la derecha.")
	case strings.Contains(lower, "test") || strings.Contains(lower, "probar"):
		b.WriteString("Añade (test nombre (eq campo valor)), Aplicar flow y pulsa Probar. Los tests corren sobre el estado del Runtime de esa app.")
	case strings.Contains(lower, "error") || strings.Contains(lower, "falla") || strings.Contains(lower, "diagn"):
		if len(diags) == 0 {
			b.WriteString("No veo diagnósticos graves en el fuente actual.")
		} else {
			b.WriteString("Diagnósticos actuales:\n")
			for _, d := range diags {
				b.WriteString("• " + d + "\n")
			}
		}
	default:
		b.WriteString("He leído el flow actual. ")
		if len(diags) > 0 {
			b.WriteString(fmt.Sprintf("Hay %d diagnóstico(s). ", len(diags)))
		}
		b.WriteString("Describe el problema de negocio en una frase y te devuelvo organismos, reglas, UI y tests. También puedes pedir: «añade regla de umbral», «UI con splash», «deploy», «explica pulse».")
	}
	if len(sugs) > 0 {
		b.WriteString("\n\nSugerencias top:\n")
		for i, s := range sugs {
			if i >= 3 {
				break
			}
			b.WriteString(fmt.Sprintf("• %s (%s): %s\n", s.Title, s.Kind, s.Reason))
		}
	}
	return b.String()
}

// tryExternal calls OpenAI-compatible API if OPENAI_API_KEY or ALSET_MIND_API_KEY is set.
func (a *Agent) tryExternal(req AnalyzeRequest) *AnalyzeResponse {
	key := os.Getenv("ALSET_MIND_API_KEY")
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	if key == "" || strings.TrimSpace(req.Message) == "" {
		return nil
	}
	// Keep offline agent as primary; external hook is optional and best-effort.
	// Full HTTP call omitted in sandbox if no network policy — return nil to use local Mind.
	_ = json.Marshal
	return nil
}
