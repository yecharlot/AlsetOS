package mind

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// FinanceStatement es el JSON canónico alset.finance.statement/1.0
type FinanceStatement struct {
	Schema string `json:"schema"`
	Entity struct {
		Name     string `json:"name"`
		NIT      string `json:"nit"`
		Period   string `json:"period"`
		Currency string `json:"currency"`
	} `json:"entity"`
	Balance struct {
		Assets struct {
			Current    float64 `json:"current"`
			NonCurrent float64 `json:"non_current"`
			Total      float64 `json:"total"`
		} `json:"assets"`
		Liabilities struct {
			Current    float64 `json:"current"`
			NonCurrent float64 `json:"non_current"`
			Total      float64 `json:"total"`
		} `json:"liabilities"`
		Equity float64 `json:"equity"`
	} `json:"balance"`
	Income struct {
		Revenue  float64 `json:"revenue"`
		COGS     float64 `json:"cogs"`
		Gross    float64 `json:"gross"`
		Opex     float64 `json:"opex"`
		EBITDA   float64 `json:"ebitda"`
		EBIT     float64 `json:"ebit"`
		Interest float64 `json:"interest"`
		Tax      float64 `json:"tax"`
		Net      float64 `json:"net"`
	} `json:"income"`
	Cashflow struct {
		Operating float64 `json:"operating"`
		Investing float64 `json:"investing"`
		Financing float64 `json:"financing"`
		Net       float64 `json:"net"`
	} `json:"cashflow"`
	Source struct {
		Files       []string `json:"files"`
		ExtractedAt string   `json:"extracted_at"`
	} `json:"source"`
}

// IngestFinance procesa kind finance-* y genera draft .aos con genes.
func (a *Agent) IngestFinance(req IngestRequest) IngestResponse {
	name := sanitizeName(req.Name)
	if name == "" {
		name = "analisis-financiero"
	}
	var warnings []string
	var st FinanceStatement

	kind := strings.ToLower(strings.TrimSpace(req.Kind))
	switch kind {
	case "finance-json", "finance":
		raw := req.Payload
		if len(raw) == 0 && req.Text != "" {
			raw = json.RawMessage(req.Text)
		}
		if err := json.Unmarshal(raw, &st); err != nil {
			return IngestResponse{
				Ok: false, Warnings: []string{"JSON financiero inválido: " + err.Error()},
				Agent: a.Name, Mode: string(ModeIngest),
			}
		}
	case "finance-pdf", "finance-xls", "finance-docx":
		// Los extractores viven en el paquete ingest/.
		// Integración recomendada:
		//   1. Guardar el archivo subido en disco temporal
		//   2. Llamar ingest.ExtractFromFile(path)
		//   3. Convertir el Statement resultante a este FinanceStatement
		//   4. Continuar con emitFinanceAOS
		//
		// Mientras el bridge no esté cableado en el handler HTTP,
		// se acepta el JSON canónico en Text/payload como fallback.
		if req.Text != "" || len(req.Payload) > 0 {
			raw := req.Payload
			if len(raw) == 0 {
				raw = json.RawMessage(req.Text)
			}
			if err := json.Unmarshal(raw, &st); err == nil {
				warnings = append(warnings, "usado JSON canónico embebido (extractor de archivo pendiente de cablear en el handler)")
				break
			}
		}
		warnings = append(warnings,
			"kind "+kind+": usa el paquete ingest/ (ExtractFromFile). "+
				"Sube el archivo al endpoint de upload o pega el JSON canónico.")
		return IngestResponse{
			Ok: false, Warnings: warnings, Agent: a.Name, Mode: string(ModeIngest),
		}
	default:
		return IngestResponse{
			Ok: false, Warnings: []string{"kind no soportado para finanzas: " + kind},
			Agent: a.Name, Mode: string(ModeIngest),
		}
	}

	if st.Schema == "" {
		st.Schema = "alset.finance.statement/1.0"
	}
	if st.Balance.Assets.Total == 0 && st.Balance.Assets.Current+st.Balance.Assets.NonCurrent > 0 {
		st.Balance.Assets.Total = st.Balance.Assets.Current + st.Balance.Assets.NonCurrent
	}
	if st.Balance.Liabilities.Total == 0 && st.Balance.Liabilities.Current+st.Balance.Liabilities.NonCurrent > 0 {
		st.Balance.Liabilities.Total = st.Balance.Liabilities.Current + st.Balance.Liabilities.NonCurrent
	}

	draft := emitFinanceAOS(name, st)
	entities := []string{"balance", "dre", "flujo", "analisis-financiero"}

	// Cálculo rápido de ratios para warnings
	if st.Balance.Liabilities.Current > 0 {
		liq := st.Balance.Assets.Current / st.Balance.Liabilities.Current
		if liq < 1.0 {
			warnings = append(warnings, fmt.Sprintf("Liquidez corriente baja: %.2f (< 1.0)", liq))
		}
	}
	if st.Balance.Assets.Total > 0 {
		end := st.Balance.Liabilities.Total / st.Balance.Assets.Total
		if end > 0.7 {
			warnings = append(warnings, fmt.Sprintf("Endeudamiento alto: %.2f (> 0.70)", end))
		}
	}

	return IngestResponse{
		Ok: true, DraftAOS: draft, Warnings: warnings, Entities: entities,
		Mode: string(ModeIngest), Agent: a.Name,
	}
}

func emitFinanceAOS(name string, st FinanceStatement) string {
	entName := st.Entity.Name
	if entName == "" {
		entName = name
	}
	period := st.Entity.Period
	if period == "" {
		period = "periodo"
	}
	// Precalcular ratios para que la UI muestre valores reales al abrir el flow
	liq := safeDiv(st.Balance.Assets.Current, st.Balance.Liabilities.Current)
	acida := safeDiv(st.Balance.Assets.Current, st.Balance.Liabilities.Current) // inventario no siempre disponible
	end := safeDiv(st.Balance.Liabilities.Total, st.Balance.Assets.Total)
	roe := safeDiv(st.Income.Net, st.Balance.Equity)
	roa := safeDiv(st.Income.Net, st.Balance.Assets.Total)
	margen := safeDiv(st.Income.Net, st.Income.Revenue)
	alerta := "ok"
	if liq < 1.0 || end > 0.7 {
		alerta = "alerta"
	}

	b := &strings.Builder{}
	fmt.Fprintf(b, `(flow %s
  (version "1.0.0")
  (desc "Análisis financiero %s — %s · panel UI Alset")
  (policy (mode model) (requires-ui true) (domain finance))

  (orge balance
    (kind ORGES)
    (cap "balance.consultar")
    (provides "balance.entity" "1.0")
    (state
      (balance.activo_corriente %.2f)
      (balance.activo_no_corriente %.2f)
      (balance.activo_total %.2f)
      (balance.pasivo_corriente %.2f)
      (balance.pasivo_no_corriente %.2f)
      (balance.pasivo_total %.2f)
      (balance.patrimonio %.2f)
      (balance.periodo %q)))

  (orge dre
    (kind ORGES)
    (cap "dre.consultar")
    (provides "dre.entity" "1.0")
    (state
      (dre.ingresos %.2f)
      (dre.costo_ventas %.2f)
      (dre.utilidad_bruta %.2f)
      (dre.gastos_operativos %.2f)
      (dre.ebitda %.2f)
      (dre.ebit %.2f)
      (dre.utilidad_neta %.2f)
      (dre.periodo %q)))

  (orge flujo
    (kind ORGES)
    (cap "flujo.consultar")
    (provides "flujo.entity" "1.0")
    (state
      (flujo.operativo %.2f)
      (flujo.inversion %.2f)
      (flujo.financiacion %.2f)
      (flujo.neto %.2f)
      (flujo.periodo %q)))

  (orge analisis-financiero
    (kind ORGES)
    (cap "finance.analizar" "finance.alertar")
    (requires "balance.entity" "1.0" "dre.entity" "1.0" "flujo.entity" "1.0")
    (provides "analisis.entity" "1.0")
    (state
      (analisis.liquidez %.4f)
      (analisis.prueba_acida %.4f)
      (analisis.endeudamiento %.4f)
      (analisis.roe %.4f)
      (analisis.roa %.4f)
      (analisis.margen %.4f)
      (analisis.vpn 0)
      (analisis.tir 0)
      (analisis.alerta %q)
      (analisis.entidad %q)
      (analisis.periodo %q))
    (gene gene.ratio.liquidez)
    (gene gene.ratio.prueba_acida)
    (gene gene.ratio.endeudamiento)
    (gene gene.ratio.roe)
    (gene gene.ratio.roa)
    (gene gene.ratio.margen_neto)
    (gene gene.vpn)
    (gene gene.tir)
    (rule liquidez-critica
      (when (lt analisis.liquidez 1.0))
      (then analisis.alerta "alerta"))
    (rule endeudamiento-alto
      (when (gt analisis.endeudamiento 0.7))
      (then analisis.alerta "alerta"))
    (goal analisis.alerta "ok"))

  (orge ui-finance
    (kind UI)
    (ui (column
      (splash %q)
      (gradient "Inteligencia financiera · organismos Alset" (style "animated"))
      (label "Estado del análisis")
      (badge (bind analisis.alerta))
      (label "Razones clave")
      (row
        (badge (bind analisis.liquidez))
        (badge (bind analisis.prueba_acida))
        (badge (bind analisis.endeudamiento)))
      (row
        (badge (bind analisis.roe))
        (badge (bind analisis.roa))
        (badge (bind analisis.margen)))
      (bottom-tabs
        (tab "Resumen" (column
          (label "Entidad")
          (badge (bind analisis.entidad))
          (label "Periodo")
          (badge (bind analisis.periodo))
          (label "Utilidad neta")
          (badge (bind dre.utilidad_neta))
          (button "Recalcular" (emit finance.recalcular))
          (button "Exportar informe" (emit finance.exportar))))
        (tab "Balance" (column
          (label "Activo corriente")
          (badge (bind balance.activo_corriente))
          (label "Activo total")
          (badge (bind balance.activo_total))
          (label "Pasivo total")
          (badge (bind balance.pasivo_total))
          (label "Patrimonio")
          (badge (bind balance.patrimonio))))
        (tab "DRE" (column
          (label "Ingresos")
          (badge (bind dre.ingresos))
          (label "Utilidad bruta")
          (badge (bind dre.utilidad_bruta))
          (label "EBITDA")
          (badge (bind dre.ebitda))
          (label "Utilidad neta")
          (badge (bind dre.utilidad_neta))))
        (tab "Flujo" (column
          (label "Operativo")
          (badge (bind flujo.operativo))
          (label "Inversión")
          (badge (bind flujo.inversion))
          (label "Financiación")
          (badge (bind flujo.financiacion))
          (label "Neto")
          (badge (bind flujo.neto))))
        (tab "Alertas" (column
          (label "Semáforo financiero")
          (badge (bind analisis.alerta))
          (label "Liquidez < 1 o endeudamiento > 0.7 dispara alerta")
          (button "Marcar revisado" (emit analisis.alerta=ok)))))))

  (assemble balance dre flujo analisis-financiero ui-finance)
  (surface ui-finance)
  (bind analisis-financiero balance balance.entity 1.0)
  (bind analisis-financiero dre dre.entity 1.0)
  (bind analisis-financiero flujo flujo.entity 1.0)
  (deploy web (devices mobile tablet desktop))
  (test ini (eq analisis.alerta %q))
)
`, name, entName, period,
		st.Balance.Assets.Current, st.Balance.Assets.NonCurrent, st.Balance.Assets.Total,
		st.Balance.Liabilities.Current, st.Balance.Liabilities.NonCurrent, st.Balance.Liabilities.Total,
		st.Balance.Equity, period,
		st.Income.Revenue, st.Income.COGS, st.Income.Gross, st.Income.Opex,
		st.Income.EBITDA, st.Income.EBIT, st.Income.Net, period,
		st.Cashflow.Operating, st.Cashflow.Investing, st.Cashflow.Financing, st.Cashflow.Net, period,
		liq, acida, end, roe, roa, margen, alerta, entName, period,
		entName+" · "+period,
		alerta)

	return b.String()
}

// Helpers numéricos usados por los genes (también en gene/finance.go)
func safeDiv(a, b float64) float64 {
	if b == 0 || math.IsNaN(b) {
		return 0
	}
	return a / b
}
