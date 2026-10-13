package ingest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ExtractFromPDF intenta extraer texto de un PDF y mapear etiquetas financieras.
// Estrategia:
//  1. Si existe pdftotext (poppler-utils) → usarlo (mejor calidad de tablas).
//  2. Fallback: lectura cruda de bytes buscando patrones de texto (muy limitada).
//  3. Siempre produce warnings claros sobre la calidad de la extracción.
//
// Recomendación de producción: pre-procesar con pdfplumber (Python) o Camelot
// y alimentar el JSON canónico. Este extractor es el “best-effort” nativo.
func ExtractFromPDF(path string) (*FinanceStatement, []string, error) {
	var warnings []string
	text, method, err := extractPDFText(path)
	if err != nil {
		return nil, []string{err.Error()}, err
	}
	warnings = append(warnings, "método de extracción: "+method)

	if len(strings.TrimSpace(text)) < 50 {
		warnings = append(warnings, "texto extraído muy corto; el PDF puede ser escaneado (necesita OCR)")
		return &FinanceStatement{Schema: "alset.finance.statement/1.0"}, warnings, nil
	}

	st, w := mapTextToStatement(text)
	warnings = append(warnings, w...)
	st.Source.Files = []string{path}
	st.Entity.Name = strings.TrimSuffix(baseName(path), ".pdf")
	return st, warnings, nil
}

func extractPDFText(path string) (string, string, error) {
	// 1. pdftotext (poppler)
	if _, err := exec.LookPath("pdftotext"); err == nil {
		out, err := exec.Command("pdftotext", "-layout", "-enc", "UTF-8", path, "-").Output()
		if err == nil && len(out) > 50 {
			return string(out), "pdftotext", nil
		}
	}

	// 2. Fallback muy básico: buscar streams de texto en el PDF (sin librería)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("leer pdf: %w", err)
	}
	// Extraer strings imprimibles largos
	re := regexp.MustCompile(`[\x20-\x7E\xC0-\xFF]{4,}`)
	matches := re.FindAll(data, -1)
	var b strings.Builder
	for _, m := range matches {
		b.Write(m)
		b.WriteByte(' ')
	}
	txt := b.String()
	if len(txt) < 50 {
		return "", "raw-fallback", fmt.Errorf("no se pudo extraer texto útil del PDF (¿escaneado?)")
	}
	return txt, "raw-fallback", nil
}

// mapTextToStatement aplica las mismas etiquetas que el extractor Excel.
func mapTextToStatement(text string) (*FinanceStatement, []string) {
	st := &FinanceStatement{Schema: "alset.finance.statement/1.0"}
	st.Entity.Currency = "COP"
	var warnings []string

	labelMap := []struct {
		keys []string
		set  func(*FinanceStatement, float64)
	}{
		{[]string{"activo corriente", "activos corrientes"}, func(s *FinanceStatement, v float64) { s.Balance.Assets.Current = v }},
		{[]string{"activo no corriente", "activos no corrientes"}, func(s *FinanceStatement, v float64) { s.Balance.Assets.NonCurrent = v }},
		{[]string{"activo total", "total activo", "total activos"}, func(s *FinanceStatement, v float64) { s.Balance.Assets.Total = v }},
		{[]string{"pasivo corriente", "pasivos corrientes"}, func(s *FinanceStatement, v float64) { s.Balance.Liabilities.Current = v }},
		{[]string{"pasivo no corriente", "pasivos no corrientes"}, func(s *FinanceStatement, v float64) { s.Balance.Liabilities.NonCurrent = v }},
		{[]string{"pasivo total", "total pasivo", "total pasivos"}, func(s *FinanceStatement, v float64) { s.Balance.Liabilities.Total = v }},
		{[]string{"patrimonio", "total patrimonio", "patrimonio total"}, func(s *FinanceStatement, v float64) { s.Balance.Equity = v }},
		{[]string{"ingresos operacionales", "ingresos netos", "ventas netas", "ingresos"}, func(s *FinanceStatement, v float64) { s.Income.Revenue = v }},
		{[]string{"costo de ventas", "costo de los bienes"}, func(s *FinanceStatement, v float64) { s.Income.COGS = v }},
		{[]string{"utilidad bruta", "margen bruto"}, func(s *FinanceStatement, v float64) { s.Income.Gross = v }},
		{[]string{"gastos operacionales", "gastos de administración", "gastos operativos"}, func(s *FinanceStatement, v float64) { s.Income.Opex = v }},
		{[]string{"ebitda"}, func(s *FinanceStatement, v float64) { s.Income.EBITDA = v }},
		{[]string{"utilidad operacional", "ebit"}, func(s *FinanceStatement, v float64) { s.Income.EBIT = v }},
		{[]string{"utilidad neta", "ganancia neta", "resultado neto"}, func(s *FinanceStatement, v float64) { s.Income.Net = v }},
		{[]string{"flujo de efectivo de las actividades de operación", "flujo operativo"}, func(s *FinanceStatement, v float64) { s.Cashflow.Operating = v }},
		{[]string{"flujo de efectivo de las actividades de inversión", "flujo de inversión"}, func(s *FinanceStatement, v float64) { s.Cashflow.Investing = v }},
		{[]string{"flujo de efectivo de las actividades de financiación", "flujo de financiación"}, func(s *FinanceStatement, v float64) { s.Cashflow.Financing = v }},
	}

	lower := strings.ToLower(text)
	// Buscar patrones: "etiqueta .... número"
	reNum := regexp.MustCompile(`(?i)([a-záéíóúñü\s]{5,40}?)\s+([\d\.\,\-\(\)]{3,})`)
	matches := reNum.FindAllStringSubmatch(lower, -1)

	found := 0
	for _, m := range matches {
		label := strings.TrimSpace(m[1])
		val := parseNumber(m[2])
		if val == 0 && !isZeroLike(m[2]) {
			continue
		}
		for _, entry := range labelMap {
			for _, k := range entry.keys {
				if strings.Contains(label, k) {
					entry.set(st, val)
					found++
					break
				}
			}
		}
	}

	// Totales derivados
	if st.Balance.Assets.Total == 0 {
		st.Balance.Assets.Total = st.Balance.Assets.Current + st.Balance.Assets.NonCurrent
	}
	if st.Balance.Liabilities.Total == 0 {
		st.Balance.Liabilities.Total = st.Balance.Liabilities.Current + st.Balance.Liabilities.NonCurrent
	}
	if st.Balance.Equity == 0 && st.Balance.Assets.Total > 0 {
		st.Balance.Equity = st.Balance.Assets.Total - st.Balance.Liabilities.Total
	}
	if st.Income.Gross == 0 && st.Income.Revenue > 0 {
		st.Income.Gross = st.Income.Revenue - st.Income.COGS
	}

	if found == 0 {
		warnings = append(warnings, "no se mapearon campos financieros; revisa el texto del PDF o usa JSON canónico")
	} else {
		warnings = append(warnings, fmt.Sprintf("mapeados ~%d campos desde texto", found))
	}
	return st, warnings
}

// ExtractFromPDFBytes versión que recibe el contenido en memoria (útil para upload).
func ExtractFromPDFBytes(name string, data []byte) (*FinanceStatement, []string, error) {
	tmp, err := os.CreateTemp("", "alset-finance-*.pdf")
	if err != nil {
		return nil, nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return nil, nil, err
	}
	tmp.Close()
	st, w, err := ExtractFromPDF(tmp.Name())
	if st != nil {
		st.Source.Files = []string{name}
		st.Entity.Name = strings.TrimSuffix(filepath.Base(name), ".pdf")
	}
	return st, w, err
}
