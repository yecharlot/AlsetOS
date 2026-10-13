package mind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIngestFinanceSampleJSON(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	sample := filepath.Join(filepath.Dir(file), "..", "ejemplos", "finance", "sample.json")
	raw, err := os.ReadFile(sample)
	if err != nil {
		t.Fatal(err)
	}
	a := NewAgent()
	resp := a.Ingest(IngestRequest{
		Kind:    "finance-json",
		Name:    "demo-pyme",
		Payload: json.RawMessage(raw),
	})
	if !resp.Ok {
		t.Fatalf("warnings=%v", resp.Warnings)
	}
	if !strings.Contains(resp.DraftAOS, "analisis-financiero") {
		t.Fatal("draft sin analisis-financiero")
	}
	if !strings.Contains(resp.DraftAOS, "balance.activo_corriente") {
		t.Fatal("draft sin balance")
	}
	if len(resp.Entities) < 3 {
		t.Fatalf("entities=%v", resp.Entities)
	}
}
