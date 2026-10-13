// CLI vertical financiero: archivo → ratios Zyrion → draft .aos
//
//	go run ./cmd/alset-finance -file ejemplos/finance/sample.json
//	go run ./cmd/alset-finance -file estados.xlsx
//	go run ./cmd/alset-finance -file sample.json -out draft.aos
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yecharlot/AlsetOS/gene"
	"github.com/yecharlot/AlsetOS/ingest"
	"github.com/yecharlot/AlsetOS/mind"
)

func main() {
	file := flag.String("file", "", "ruta a .json .xlsx .xls .pdf .docx")
	out := flag.String("out", "", "escribir draft .aos aquí (opcional)")
	name := flag.String("name", "", "nombre del flow (default: del archivo)")
	flag.Parse()
	if *file == "" {
		fmt.Fprintf(os.Stderr, "uso: alset-finance -file <ruta> [-out draft.aos] [-name analisis]\n")
		os.Exit(2)
	}

	res, err := ingest.ExtractFromFile(*file)
	if err != nil && (res == nil || res.Statement == nil) {
		fmt.Fprintf(os.Stderr, "error extracción: %v\n", err)
		os.Exit(1)
	}
	st := res.Statement
	for _, w := range res.Warnings {
		fmt.Println("⚠️ ", w)
	}
	if err != nil {
		fmt.Println("⚠️ ", err)
	}

	fmt.Println("=== Alset Finance ===")
	fmt.Printf("Fuente: %s\n", res.Source)
	fmt.Printf("Entidad: %s | Periodo: %s | Moneda: %s\n",
		st.Entity.Name, st.Entity.Period, st.Entity.Currency)

	liq := gene.LiquidezCorriente(st.Balance.Assets.Current, st.Balance.Liabilities.Current)
	acida := gene.PruebaAcida(st.Balance.Assets.Current, 0, st.Balance.Liabilities.Current)
	end := gene.Endeudamiento(st.Balance.Liabilities.Total, st.Balance.Assets.Total)
	roe := gene.ROE(st.Income.Net, st.Balance.Equity)
	roa := gene.ROA(st.Income.Net, st.Balance.Assets.Total)
	margen := gene.MargenNeto(st.Income.Net, st.Income.Revenue)
	alerta := gene.AlertaFinanciera(liq.Value, acida.Value, end.Value, roe.Value, roa.Value, margen.Value)

	printRatio := func(label string, r gene.RatioResult) {
		z := map[int]string{0: "NO", 1: "SI", 2: "INCIERTO"}[r.Zyrion]
		fmt.Printf("  %-16s %10.4f  Zyrion=%s  (%s)\n", label, r.Value, z, r.Message)
	}
	fmt.Println("\n--- Razones financieras (genes) ---")
	printRatio("Liquidez", liq)
	printRatio("Prueba ácida", acida)
	printRatio("Endeudamiento", end)
	printRatio("ROE", roe)
	printRatio("ROA", roa)
	printRatio("Margen neto", margen)
	printRatio("Alerta global", alerta)

	// VPN/TIR de ejemplo con flujo operativo simplificado
	flujos := []float64{
		-st.Balance.Assets.Total * 0.1,
		st.Cashflow.Operating * 0.25,
		st.Cashflow.Operating * 0.25,
		st.Cashflow.Operating * 0.25,
		st.Cashflow.Operating * 0.25,
	}
	vpn := gene.VPN(flujos, 0.12)
	tir := gene.TIR(flujos)
	fmt.Println("\n--- VPN / TIR (flujos ilustrativos) ---")
	printRatio("VPN@12%", vpn)
	printRatio("TIR", tir)

	n := *name
	if n == "" {
		n = strings.TrimSuffix(filepath.Base(*file), filepath.Ext(*file))
		n = mindSanitize(n)
		if n == "" {
			n = "analisis-financiero"
		}
	}

	raw, _ := json.Marshal(st)
	agent := mind.NewAgent()
	ing := agent.Ingest(mind.IngestRequest{
		Kind:    "finance-json",
		Name:    n,
		Payload: raw,
	})
	if !ing.Ok {
		fmt.Println("\nIngest Mind falló:", strings.Join(ing.Warnings, "; "))
		os.Exit(1)
	}
	fmt.Println("\n--- Draft .aos (Mind IngestFinance) ---")
	fmt.Printf("Entidades: %v\n", ing.Entities)
	for _, w := range ing.Warnings {
		fmt.Println("⚠️ ", w)
	}
	if *out != "" {
		if err := os.WriteFile(*out, []byte(ing.DraftAOS), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "no se pudo escribir %s: %v\n", *out, err)
			os.Exit(1)
		}
		fmt.Println("Escrito:", *out)
	} else {
		// preview corto
		lines := strings.Split(ing.DraftAOS, "\n")
		max := 40
		if len(lines) < max {
			max = len(lines)
		}
		fmt.Println(strings.Join(lines[:max], "\n"))
		if len(lines) > max {
			fmt.Printf("... (%d líneas más; usa -out para guardar completo)\n", len(lines)-max)
		}
	}
}

func mindSanitize(s string) string {
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
