package ingest

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ExtractFromCSV lee un CSV de dos columnas etiqueta,valor (o key,value).
func ExtractFromCSV(path string) (*FinanceStatement, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.TrimLeadingSpace = true
	rows, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	st := &FinanceStatement{Schema: "alset.finance.statement/1.0"}
	st.Entity.Currency = "USD"
	st.Source.Files = []string{path}
	var warnings []string
	setters := map[string]func(float64){
		"activo_corriente":        func(v float64) { st.Balance.Assets.Current = v },
		"activo corriente":        func(v float64) { st.Balance.Assets.Current = v },
		"activo_no_corriente":     func(v float64) { st.Balance.Assets.NonCurrent = v },
		"activo total":            func(v float64) { st.Balance.Assets.Total = v },
		"pasivo_corriente":        func(v float64) { st.Balance.Liabilities.Current = v },
		"pasivo corriente":        func(v float64) { st.Balance.Liabilities.Current = v },
		"pasivo_no_corriente":     func(v float64) { st.Balance.Liabilities.NonCurrent = v },
		"pasivo total":            func(v float64) { st.Balance.Liabilities.Total = v },
		"patrimonio":              func(v float64) { st.Balance.Equity = v },
		"equity":                  func(v float64) { st.Balance.Equity = v },
		"ingresos":                func(v float64) { st.Income.Revenue = v },
		"revenue":                 func(v float64) { st.Income.Revenue = v },
		"ventas":                  func(v float64) { st.Income.Revenue = v },
		"cogs":                    func(v float64) { st.Income.COGS = v },
		"costo de ventas":         func(v float64) { st.Income.COGS = v },
		"gastos":                  func(v float64) { st.Income.Opex = v },
		"opex":                    func(v float64) { st.Income.Opex = v },
		"utilidad neta":           func(v float64) { st.Income.Net = v },
		"net_income":              func(v float64) { st.Income.Net = v },
		"flujo operativo":         func(v float64) { st.Cashflow.Operating = v },
		"flujo inversion":         func(v float64) { st.Cashflow.Investing = v },
		"flujo financiacion":      func(v float64) { st.Cashflow.Financing = v },
	}
	for i, row := range rows {
		if len(row) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(row[0]))
		val := strings.ReplaceAll(strings.TrimSpace(row[1]), ",", "")
		val = strings.ReplaceAll(val, " ", "")
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			if i == 0 {
				continue // header
			}
			warnings = append(warnings, fmt.Sprintf("fila %d no numérica: %s", i+1, row[1]))
			continue
		}
		if fn, ok := setters[key]; ok {
			fn(n)
		} else {
			warnings = append(warnings, "etiqueta no mapeada: "+key)
		}
	}
	if st.Balance.Assets.Total == 0 {
		st.Balance.Assets.Total = st.Balance.Assets.Current + st.Balance.Assets.NonCurrent
	}
	if st.Balance.Liabilities.Total == 0 {
		st.Balance.Liabilities.Total = st.Balance.Liabilities.Current + st.Balance.Liabilities.NonCurrent
	}
	if st.Income.Gross == 0 && st.Income.Revenue > 0 {
		st.Income.Gross = st.Income.Revenue - st.Income.COGS
	}
	return st, warnings, nil
}
