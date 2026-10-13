package mind

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// WorkMode define cómo opera el Studio / Inspector.
// No todo flujo termina en pantallas: análisis y modelado son válidos.
type WorkMode string

const (
	ModeAnalyze    WorkMode = "analyze"    // Auditar dominio → ruta de digitalización (sin UI obligatoria)
	ModeModel      WorkMode = "model"      // Solo ORGES de dominio + rules + binds + tests
	ModeIngest     WorkMode = "ingest"     // Schema / SQL / OpenAPI → draft .aos
	ModeImplement  WorkMode = "implement"  // Dominio + UI opcional + deploy
	ModeAudit      WorkMode = "audit"      // Sombra PrismaTEC / tokenización / Pulse
)

func ParseWorkMode(s string) WorkMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "analyze", "analisis", "análisis":
		return ModeAnalyze
	case "model", "modelo":
		return ModeModel
	case "ingest", "ingesta", "schema":
		return ModeIngest
	case "audit", "auditoria", "auditoría", "token":
		return ModeAudit
	default:
		return ModeImplement
	}
}

// ModeInfo describe el modo para la UI del Studio.
type ModeInfo struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Summary     string   `json:"summary"`
	RequiresUI  bool     `json:"requires_ui"`
	Outputs     []string `json:"outputs"`
	NextSteps   []string `json:"next_steps"`
}

func CatalogModes() []ModeInfo {
	return []ModeInfo{
		{
			ID: string(ModeAnalyze), Label: "Análisis",
			Summary: "Describe el negocio. Obtén diagnóstico, riesgos y ruta de digitalización. No genera pantallas.",
			RequiresUI: false,
			Outputs:    []string{"roadmap", "entidades candidatas", "brechas", "draft .aos opcional sin UI"},
			NextSteps:  []string{"Pasar a Modelo", "Priorizar un subdominio piloto"},
		},
		{
			ID: string(ModeModel), Label: "Modelo ORGES",
			Summary: "Traduce el dominio a organismos (state, cap, rule, bind, test) listos para validar — sin surface visual.",
			RequiresUI: false,
			Outputs:    []string{".aos de dominio", "contratos", "tests"},
			NextSteps:  []string{"Probar tests", "Añadir UI solo si el cliente la pide", "Modo Auditoría"},
		},
		{
			ID: string(ModeIngest), Label: "Ingestión",
			Summary: "Importa esquema BD, SQL, API REST o estados financieros (PDF/XLS/DOCX/JSON) y genera draft ORGES + genes.",
			RequiresUI: false,
			Outputs:    []string{"IR intermedio", "draft .aos", "ratios/TIR/VPN", "warnings"},
			NextSteps:  []string{"Revisar en Modelo", "Firmar manifiesto", "Implementar"},
		},
		{
			ID: string(ModeImplement), Label: "Implementación",
			Summary: "ORGES + superficie UI opcional + preview + deploy multi-dispositivo.",
			RequiresUI: true,
			Outputs:    []string{"app desplegable", "preview", "tests"},
			NextSteps:  []string{"Piloto con usuario real"},
		},
		{
			ID: string(ModeAudit), Label: "Auditoría / Tokenización",
			Summary: "Prepara el reflejo para red PrismaTEC: Pulse firmado, hash chain, evidencia de mutaciones y anclaje del modelo.",
			RequiresUI: false,
			Outputs:    []string{"policy shadow", "hooks Pulse", "fingerprint del .aos", "notas de tokenización"},
			NextSteps:  []string{"Conectar nodo PrismaTec-Core", "Modo shadow antes de enforce"},
		},
	}
}

// RoadmapFromDomain genera análisis de digitalización sin exigir UI.
func (a *Agent) RoadmapFromDomain(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "Describe el negocio: actores, procesos, datos críticos, dolores y sistemas actuales."
	}
	l := strings.ToLower(msg)
	var b strings.Builder
	b.WriteString("## Ruta de digitalización (modo Análisis)\n\n")
	b.WriteString("**Dominio declarado:** " + truncate(msg, 280) + "\n\n")
	b.WriteString("### 1. Entidades candidatas (ORGES)\n")
	ents := guessEntities(l)
	for _, e := range ents {
		b.WriteString("- `" + e + "` — state + capacidades consultar/actualizar\n")
	}
	b.WriteString("\n### 2. Procesos a modelar primero\n")
	b.WriteString("- Flujo feliz del núcleo del negocio (crear → confirmar → cerrar)\n")
	b.WriteString("- Puntos de control (stock, caja, estados, roles)\n")
	b.WriteString("- Evidencia: qué mutaciones deben firmarse (Pulse)\n\n")
	b.WriteString("### 3. Sistemas legados a reflejar\n")
	if strings.Contains(l, "excel") || strings.Contains(l, "papel") {
		b.WriteString("- Hoy: planillas/papel → ingestión CSV/JSON en modo Ingestión\n")
	} else if strings.Contains(l, "sql") || strings.Contains(l, "base de datos") || strings.Contains(l, "postgres") {
		b.WriteString("- Hoy: BD relacional → conector schema → draft ORGES\n")
	} else {
		b.WriteString("- Identificar: BD, APIs, archivos; usar modo Ingestión\n")
	}
	b.WriteString("\n### 4. Fases recomendadas\n")
	b.WriteString("1. **Modelo** — ORGES + rules + tests (sin pantallas)\n")
	b.WriteString("2. **Piloto** — un subproceso real\n")
	b.WriteString("3. **Implementación** — UI solo donde aporte valor\n")
	b.WriteString("4. **Auditoría** — sombra PrismaTEC / tokenización del modelo\n\n")
	b.WriteString("### 5. Qué no digitalizar aún\n")
	b.WriteString("- Reportes complejos, integraciones periféricas, automatizaciones opacas\n")
	b.WriteString("- Cualquier regla que no puedas expresar como predicado o contrato claro\n")
	return b.String()
}

func guessEntities(l string) []string {
	cands := []struct {
		k string
		e string
	}{
		{"pedido", "pedidos"}, {"venta", "ventas"}, {"stock", "inventario"}, {"inventario", "inventario"},
		{"cliente", "clientes"}, {"caja", "caja"}, {"pago", "pagos"}, {"cita", "citas"},
		{"empleado", "personal"}, {"nómina", "nomina"}, {"factura", "facturacion"},
		{"producto", "productos"}, {"ruta", "logistica"}, {"usuario", "auth"},
	}
	seen := map[string]bool{}
	var out []string
	for _, c := range cands {
		if strings.Contains(l, c.k) && !seen[c.e] {
			seen[c.e] = true
			out = append(out, c.e)
		}
	}
	if len(out) == 0 {
		out = []string{"nucleo", "politica", "auditoria"}
	}
	return out
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// BuildDomainOnly genera .aos sin UI (modos analyze/model).
func (a *Agent) BuildDomainOnly(name, msg string) string {
	name = sanitizeName(name)
	if name == "" {
		name = "modelo-dominio"
	}
	ents := guessEntities(strings.ToLower(msg))
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`(flow %s
  (version "1.0.0")
  (desc %q)
  (policy (mode model) (requires-ui false))
`, name, truncate(msg, 120)))
	var assemble []string
	for _, e := range ents {
		assemble = append(assemble, e)
		b.WriteString(fmt.Sprintf(`  (orge %s
    (kind ORGES)
    (cap "%s.consultar" "%s.actualizar")
    (provides "%s.entity" "1.0")
    (state (%s.activo true) (%s.alerta false))
    (rule %s-ok (when (eq %s.activo true)) (then %s.alerta false))
    (goal %s.alerta false))
`, e, e, e, e, e, e, e, e, e, e))
	}
	// simple bind chain if 2+
	if len(ents) >= 2 {
		b.WriteString(fmt.Sprintf("  (bind %s %s %s.entity 1.0)\n", ents[1], ents[0], ents[0]))
	}
	b.WriteString("  (assemble " + strings.Join(assemble, " ") + ")\n")
	if len(ents) > 0 {
		b.WriteString(fmt.Sprintf("  (test ini (eq %s.alerta false))\n", ents[0]))
	}
	b.WriteString(")\n")
	return b.String()
}

func sanitizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else if r == ' ' || r == '_' {
			b.WriteByte('-')
		}
	}
	return b.String()
}

// --- Ingestión: schema JSON / SQL / OpenAPI-ish ---

type IngestRequest struct {
	Kind    string          `json:"kind"` // schema | sql | openapi | rest
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
	Text    string          `json:"text"` // SQL crudo u OpenAPI yaml/json texto
	Mode    string          `json:"mode"`
}

type IngestResponse struct {
	Ok       bool     `json:"ok"`
	DraftAOS string   `json:"draft_aos"`
	Warnings []string `json:"warnings"`
	Entities []string `json:"entities"`
	Mode     string   `json:"mode"`
	Agent    string   `json:"agent"`
}

// Intermediate field from schema
type schemaField struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Nullable bool     `json:"nullable"`
	PK       bool     `json:"pk"`
	Checks   []string `json:"checks"`
}

type schemaEntity struct {
	Name   string        `json:"name"`
	Fields []schemaField `json:"fields"`
}

type schemaDoc struct {
	Entities []schemaEntity `json:"entities"`
	Tables   []schemaEntity `json:"tables"` // alias
}

func (a *Agent) Ingest(req IngestRequest) IngestResponse {
	name := sanitizeName(req.Name)
	if name == "" {
		name = "ingest-draft"
	}
	kind := strings.ToLower(strings.TrimSpace(req.Kind))
	if kind == "" {
		kind = "schema"
	}
	// Vertical financiero: PDF/XLS/DOCX/JSON → ORGES + genes
	switch kind {
	case "finance-json", "finance", "finance-pdf", "finance-xls", "finance-docx":
		return a.IngestFinance(req)
	}
	var warnings []string
	var entities []schemaEntity

	switch kind {
	case "sql":
		entities, warnings = parseSQLCreate(req.Text)
	case "openapi", "rest":
		entities, warnings = parseRESTHints(req.Text)
	default: // schema JSON
		var doc schemaDoc
		raw := req.Payload
		if len(raw) == 0 && req.Text != "" {
			raw = json.RawMessage(req.Text)
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			// try array of entities
			var arr []schemaEntity
			if err2 := json.Unmarshal(raw, &arr); err2 != nil {
				return IngestResponse{Ok: false, Warnings: []string{"JSON de esquema inválido: " + err.Error()}, Agent: a.Name}
			}
			entities = arr
		} else {
			entities = doc.Entities
			if len(entities) == 0 {
				entities = doc.Tables
			}
		}
	}

	if len(entities) == 0 {
		warnings = append(warnings, "No se detectaron entidades; se emite orge placeholder")
		entities = []schemaEntity{{Name: "nucleo", Fields: []schemaField{{Name: "id", Type: "string", PK: true}}}}
	}

	draft := emitAOSFromSchema(name, entities, warnings)
	names := make([]string, 0, len(entities))
	for _, e := range entities {
		names = append(names, sanitizeName(e.Name))
	}
	return IngestResponse{
		Ok: true, DraftAOS: draft, Warnings: warnings, Entities: names,
		Mode: string(ModeIngest), Agent: a.Name,
	}
}

var reCreateTable = regexp.MustCompile(`(?i)create\s+table\s+(?:if\s+not\s+exists\s+)?["'\x60]?(\w+)["'\x60]?\s*\(([\s\S]*?)\)\s*;?`)

func parseSQLCreate(sql string) ([]schemaEntity, []string) {
	var warnings []string
	var out []schemaEntity
	matches := reCreateTable.FindAllStringSubmatch(sql, -1)
	if len(matches) == 0 {
		warnings = append(warnings, "No se encontraron CREATE TABLE; pega DDL completo")
		return out, warnings
	}
	for _, m := range matches {
		table := m[1]
		body := m[2]
		ent := schemaEntity{Name: table}
		for _, line := range strings.Split(body, ",") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(strings.ToUpper(line), "PRIMARY") ||
				strings.HasPrefix(strings.ToUpper(line), "FOREIGN") ||
				strings.HasPrefix(strings.ToUpper(line), "CONSTRAINT") ||
				strings.HasPrefix(strings.ToUpper(line), "UNIQUE") ||
				strings.HasPrefix(strings.ToUpper(line), "KEY") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) < 1 {
				continue
			}
			fname := strings.Trim(parts[0], "\"'`")
			ftype := "string"
			if len(parts) > 1 {
				ftype = mapSQLType(parts[1])
			}
			pk := strings.Contains(strings.ToUpper(line), "PRIMARY KEY")
			f := schemaField{Name: fname, Type: ftype, PK: pk, Nullable: !strings.Contains(strings.ToUpper(line), "NOT NULL")}
			if strings.Contains(strings.ToUpper(line), "CHECK") {
				f.Checks = append(f.Checks, "legacy-check")
				warnings = append(warnings, table+"."+fname+": CHECK detectado — revisar rule ARL manualmente")
			}
			ent.Fields = append(ent.Fields, f)
		}
		out = append(out, ent)
	}
	return out, warnings
}

func mapSQLType(t string) string {
	u := strings.ToUpper(t)
	switch {
	case strings.Contains(u, "INT"), strings.Contains(u, "SERIAL"):
		return "int"
	case strings.Contains(u, "BOOL"):
		return "bool"
	case strings.Contains(u, "DECIMAL"), strings.Contains(u, "NUMERIC"), strings.Contains(u, "REAL"), strings.Contains(u, "FLOAT"):
		return "decimal"
	case strings.Contains(u, "DATE"), strings.Contains(u, "TIME"):
		return "ts"
	default:
		return "string"
	}
}

func parseRESTHints(text string) ([]schemaEntity, []string) {
	var warnings []string
	// Very light: look for /resources and path segments
	rePath := regexp.MustCompile(`(?i)(/[a-z0-9_\-{}]+)`)
	found := map[string]bool{}
	var ents []schemaEntity
	for _, m := range rePath.FindAllString(text, -1) {
		seg := strings.Trim(m, "/")
		seg = strings.Split(seg, "/")[0]
		seg = strings.Trim(seg, "{}")
		if seg == "" || seg == "api" || seg == "v1" || seg == "v2" {
			continue
		}
		if found[seg] {
			continue
		}
		found[seg] = true
		ents = append(ents, schemaEntity{
			Name: seg,
			Fields: []schemaField{
				{Name: "id", Type: "string", PK: true},
				{Name: "payload", Type: "string", Nullable: true},
			},
		})
	}
	if len(ents) == 0 {
		warnings = append(warnings, "No se extrajeron recursos REST; describe paths o pega OpenAPI")
	} else {
		warnings = append(warnings, "Ingestión REST heurística: revisar caps y contratos a mano")
	}
	return ents, warnings
}

func emitAOSFromSchema(flowName string, entities []schemaEntity, warnings []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`(flow %s
  (version "1.0.0")
  (desc "Draft generado por ingestión de esquema — revisar antes de producción")
  (policy (mode ingest) (requires-ui false) (status draft))
`, flowName))
	var names []string
	for _, e := range entities {
		n := sanitizeName(e.Name)
		if n == "" {
			continue
		}
		names = append(names, n)
		b.WriteString(fmt.Sprintf("  (orge %s\n    (kind ORGES)\n", n))
		b.WriteString(fmt.Sprintf("    (cap \"%s.consultar\" \"%s.actualizar\")\n", n, n))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", n))
		b.WriteString("    (state")
		if len(e.Fields) == 0 {
			b.WriteString(fmt.Sprintf(" (%s.id \"\")", n))
		}
		for _, f := range e.Fields {
			fn := sanitizeName(f.Name)
			if fn == "" {
				continue
			}
			def := defaultForType(f.Type)
			b.WriteString(fmt.Sprintf(" (%s.%s %s)", n, fn, def))
		}
		b.WriteString(fmt.Sprintf(" (%s.alerta false))\n", n))
		b.WriteString(fmt.Sprintf("    (goal %s.alerta false))\n", n))
	}
	if len(names) >= 2 {
		b.WriteString(fmt.Sprintf("  (bind %s %s %s.entity 1.0)\n", names[1], names[0], names[0]))
	}
	if len(names) > 0 {
		b.WriteString("  (assemble " + strings.Join(names, " ") + ")\n")
		b.WriteString(fmt.Sprintf("  (test ingest-ini (eq %s.alerta false))\n", names[0]))
	}
	for _, w := range warnings {
		b.WriteString(fmt.Sprintf("  ;; warning: %s\n", w))
	}
	b.WriteString(")\n")
	return b.String()
}

func defaultForType(t string) string {
	switch strings.ToLower(t) {
	case "int", "decimal":
		return "0"
	case "bool":
		return "false"
	default:
		return `""`
	}
}

// AuditNotes texto para modo auditoría / tokenización (PrismaTEC).
func (a *Agent) AuditNotes(flowName, src string) string {
	fp := fmt.Sprintf("%d-chars", len(src))
	return fmt.Sprintf(`## Modo Auditoría / Tokenización

Flow: %s
Huella textual del manifiesto: %s

### Política recomendada (PrismaTEC-Core)
1. **shadow** — observar mutaciones del sistema actual; firmar Pulse (Ed25519 + nonce).
2. **alert** — divergencia entre BD legada y state ORGES.
3. **enforce** — solo cuando tests y piloto sean estables.

### Tokenización del modelo (no del hype)
- El activo no es un token especulativo: es el **manifiesto .aos versionado + cadena de evidencia**.
- Cada versión firmada del flow puede anclarse (hash) en nodos PrismaTEC.
- Capacidades (cap) y contratos (bind) son la interfaz gobernable del negocio.

### Próximos pasos técnicos
- Desplegar nodo PrismaTec-Core (organism + pulse).
- Agente que emita Pulse en cada cap del runtime Alset.
- Checkpoint diario de la hash chain.

Referencia repos: PrismaTec-Core (organism, pulse), AlsetOS (runtime ORGES).
`, flowName, fp)
}
