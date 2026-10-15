package mind

import (
	"fmt"
	"strings"
)

// InterviewInput captura la entrevista técnica del Arquitecto.
type InterviewInput struct {
	Empresa       string `json:"empresa"`
	Sector        string `json:"sector"`
	Procesos      string `json:"procesos"`
	Funcionales   string `json:"funcionales"`
	NoFuncionales string `json:"nfuncionales"`
	Sistemas      string `json:"sistemas"`
	Dolores       string `json:"dolores"`
	WorkMode      string `json:"work_mode"`
}

// ArchitectReport es el entregable completo: resumen, mejoras, estructura, ORGES, siguiente paso comercial.
type ArchitectReport struct {
	Ok           bool     `json:"ok"`
	Title        string   `json:"title"`
	Summary      string   `json:"summary"`
	Structure    []string `json:"structure"`
	Improvements []string `json:"improvements"`
	Risks        []string `json:"risks"`
	OrgesPlan    []string `json:"orges_plan"`
	DraftAOS     string   `json:"draft_aos"`
	Roadmap      string   `json:"roadmap"`
	Commercial   string   `json:"commercial"`
	Warnings     []string `json:"warnings"`
}

// AnalyzeInterview produce informe + draft ORGES a partir de la entrevista.
func (a *Agent) AnalyzeInterview(in InterviewInput) ArchitectReport {
	if in.Sector == "" {
		in.Sector = "servicios"
	}
	if in.Empresa == "" {
		in.Empresa = "negocio"
	}
	name := sanitizeName(in.Empresa)
	procs := splitList(in.Procesos)
	if len(procs) == 0 {
		procs = []string{"operacion"}
	}
	funcs := in.Funcionales
	nfuncs := in.NoFuncionales
	sistemas := in.Sistemas
	dolores := in.Dolores

	structLines := []string{
		fmt.Sprintf("Organización: %s", in.Empresa),
		fmt.Sprintf("Sector: %s", in.Sector),
		fmt.Sprintf("Procesos detectados (%d): %s", len(procs), strings.Join(procs, ", ")),
	}
	if sistemas != "" {
		structLines = append(structLines, "Sistemas actuales: "+sistemas)
	}
	if funcs != "" {
		structLines = append(structLines, "RF: "+truncStr(funcs, 240))
	}
	if nfuncs != "" {
		structLines = append(structLines, "RNF: "+truncStr(nfuncs, 240))
	}

	impr := []string{}
	risks := []string{}
	blob := strings.ToLower(funcs + " " + nfuncs + " " + dolores + " " + sistemas + " " + strings.Join(procs, " "))
	if strings.Contains(blob, "excel") || strings.Contains(blob, "papel") {
		impr = append(impr, "Migrar registros de Excel/papel a estado ORGES versionado (evita pérdida y duplicados).")
	}
	if strings.Contains(blob, "stock") || strings.Contains(blob, "inventario") {
		impr = append(impr, "Organismo de inventario con reglas ARL de stock mínimo y alerta.")
		risks = append(risks, "Quiebres de stock no detectados a tiempo.")
	}
	if strings.Contains(blob, "venta") || strings.Contains(blob, "caja") || strings.Contains(blob, "pos") {
		impr = append(impr, "Organismo de ventas/caja con emits de cobro y traza auditables.")
	}
	if strings.Contains(blob, "offline") || strings.Contains(blob, "sin internet") {
		impr = append(impr, "Priorizar modo offline-first y sincronización posterior (Alset + CID).")
	}
	if strings.Contains(blob, "multi") || strings.Contains(blob, "usuario") || strings.Contains(blob, "rol") {
		impr = append(impr, "Separar organismo de identidad/roles con capacidades (cap) por perfil.")
	}
	if strings.Contains(blob, "factura") || strings.Contains(blob, "contab") {
		impr = append(impr, "Enlazar con vertical contable (ÁbacoPhy) o genes financieros.")
	}
	if len(impr) == 0 {
		impr = append(impr, "Definir 3 procesos prioritarios y convertirlos en ORGES con tests antes de UI.")
	}
	if dolores != "" {
		risks = append(risks, "Dolores declarados: "+truncStr(dolores, 200))
	}
	if len(risks) == 0 {
		risks = append(risks, "Alcance difuso: acotar un piloto de 2 semanas en un solo proceso.")
	}

	orgesPlan := []string{"orge contexto (empresa, sector, RNF)"}
	for _, p := range procs {
		orgesPlan = append(orgesPlan, "orge "+sanitizeName(p)+" (proceso operativo)")
	}
	orgesPlan = append(orgesPlan, "contratos bind entre procesos relacionados")
	orgesPlan = append(orgesPlan, "tests iniciales de estado")
	if in.WorkMode == "implement" || in.WorkMode == "implementacion" {
		orgesPlan = append(orgesPlan, "orge UI surface + deploy multi-dispositivo")
	}

	draft := a.buildInterviewAOS(name, in, procs)
	roadmap := a.RoadmapFromDomain(fmt.Sprintf("%s %s %s %s", in.Empresa, in.Sector, funcs, dolores))
	summary := fmt.Sprintf(
		"%s opera en %s con foco en %s. La entrevista sugiere un modelo ORGES de %d organismos de proceso más contexto, "+
			"con mejoras orientadas a trazabilidad y reglas ARL. Modo de trabajo recomendado: partir por **Modelo** (sin UI) y solo luego Implementación.",
		in.Empresa, in.Sector, strings.Join(procs[:min(3, len(procs))], ", "), len(procs),
	)
	commercial := `## Oferta comercial sugerida (SaaS / servicio)
1. **Diagnóstico Arquitecto** (esta entrevista + informe): entrega resumen, riesgos y draft .aos.
2. **Modelo ORGES piloto**: 1–3 procesos tokenizados (CID + cadena de evidencia).
3. **Implementación opcional**: UI/deploy solo si el cliente lo necesita.
4. **Auditoría continua**: anclar versiones del manifiesto y eventos de negocio (token Alset, no coin especulativo).

Producto: "Arquitecto de negocios con organismos digitales" — vendes claridad y gobernanza, no solo pantallas.`

	return ArchitectReport{
		Ok: true,
		Title: fmt.Sprintf("Informe Arquitecto — %s", in.Empresa),
		Summary: summary,
		Structure: structLines,
		Improvements: impr,
		Risks: risks,
		OrgesPlan: orgesPlan,
		DraftAOS: draft,
		Roadmap: roadmap,
		Commercial: commercial,
	}
}

func (a *Agent) buildInterviewAOS(name string, in InterviewInput, procs []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("(flow %s\n", name))
	b.WriteString("  (version \"1.0.0\")\n")
	b.WriteString(fmt.Sprintf("  (desc %q)\n", "Modelo "+in.Empresa+" · "+in.Sector+" — generado por entrevista Arquitecto"))
	mode := "model"
	if in.WorkMode == "implement" || in.WorkMode == "implementacion" {
		mode = "implement"
	}
	b.WriteString(fmt.Sprintf("  (policy (mode %s) (domain %q))\n\n", mode, in.Sector))
	b.WriteString("  (orge contexto\n")
	b.WriteString("    (kind ORGES)\n")
	b.WriteString("    (cap \"dominio.consultar\")\n")
	b.WriteString("    (provides \"contexto.entity\" \"1.0\")\n")
	b.WriteString("    (state\n")
	b.WriteString(fmt.Sprintf("      (contexto.empresa %q)\n", in.Empresa))
	b.WriteString(fmt.Sprintf("      (contexto.sector %q)\n", in.Sector))
	b.WriteString(fmt.Sprintf("      (contexto.sistemas %q)\n", truncStr(in.Sistemas, 120)))
	b.WriteString(fmt.Sprintf("      (contexto.nfunc %q))\n", truncStr(in.NoFuncionales, 120)))
	b.WriteString("  )\n\n")
	assemble := []string{"contexto"}
	for i, pr := range procs {
		if i >= 8 {
			break
		}
		oid := sanitizeName(pr)
		if oid == "" {
			oid = fmt.Sprintf("proc%d", i)
		}
		assemble = append(assemble, oid)
		b.WriteString(fmt.Sprintf("  (orge %s\n", oid))
		b.WriteString("    (kind ORGES)\n")
		b.WriteString(fmt.Sprintf("    (cap \"%s.operar\" \"%s.consultar\")\n", oid, oid))
		b.WriteString(fmt.Sprintf("    (provides \"%s.entity\" \"1.0\")\n", oid))
		b.WriteString(fmt.Sprintf("    (state (%s.activo true) (%s.eventos 0))\n", oid, oid))
		pl := strings.ToLower(pr + " " + in.Funcionales)
		if strings.Contains(pl, "stock") || strings.Contains(pl, "inventario") {
			b.WriteString(fmt.Sprintf("    (state (%s.stock 10) (%s.alerta false))\n", oid, oid))
			b.WriteString(fmt.Sprintf("    (rule alerta-stock (when (lt %s.stock 3)) (then %s.alerta true))\n", oid, oid))
		}
		if strings.Contains(pl, "venta") || strings.Contains(pl, "caja") {
			b.WriteString(fmt.Sprintf("    (state (%s.ventas 0))\n", oid))
		}
		b.WriteString("  )\n\n")
	}
	b.WriteString("  (assemble " + strings.Join(assemble, " ") + ")\n")
	b.WriteString(fmt.Sprintf("  (test ini (eq contexto.empresa %q))\n", in.Empresa))
	b.WriteString(")\n")
	return b.String()
}

func splitList(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '|'
	})
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func truncStr(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
