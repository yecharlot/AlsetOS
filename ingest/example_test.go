package ingest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestExtractFromJSON verifica el camino feliz con JSON canónico.
func TestExtractFromJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.json")
	sample := `{
  "schema": "alset.finance.statement/1.0",
  "entity": {"name": "Acme SAS", "period": "2025-Q3", "currency": "COP"},
  "balance": {
    "assets": {"current": 150000000, "non_current": 320000000, "total": 470000000},
    "liabilities": {"current": 80000000, "non_current": 120000000, "total": 200000000},
    "equity": 270000000
  },
  "income": {
    "revenue": 410000000, "cogs": 180000000, "gross": 230000000,
    "opex": 95000000, "ebitda": 135000000, "ebit": 120000000,
    "interest": 15000000, "tax": 28000000, "net": 77000000
  },
  "cashflow": {"operating": 110000000, "investing": -45000000, "financing": -20000000, "net": 45000000}
}`
	if err := os.WriteFile(path, []byte(sample), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := ExtractFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if res.Statement.Income.Net != 77000000 {
		t.Fatalf("utilidad neta esperada 77000000, got %v", res.Statement.Income.Net)
	}
	if res.Source != "json" {
		t.Fatalf("source=%s", res.Source)
	}

	b, err := res.ToCanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var st FinanceStatement
	if err := json.Unmarshal(b, &st); err != nil {
		t.Fatal(err)
	}
	if st.Balance.Assets.Total != 470000000 {
		t.Fatalf("activo total %v", st.Balance.Assets.Total)
	}
}
