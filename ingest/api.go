package ingest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExtractResult es la salida unificada de cualquier extractor.
type ExtractResult struct {
	Statement *FinanceStatement `json:"statement"`
	Warnings  []string          `json:"warnings"`
	Source    string            `json:"source"` // pdf | xls | docx | json
}

// ExtractFromFile detecta la extensión y llama al extractor correspondiente.
func ExtractFromFile(path string) (*ExtractResult, error) {
	ext := strings.ToLower(filepath.Ext(path))
	var st *FinanceStatement
	var warnings []string
	var err error
	source := ""

	switch ext {
	case ".xlsx", ".xls":
		st, warnings, err = ExtractFromXLS(path)
		source = "xls"
	case ".pdf":
		st, warnings, err = ExtractFromPDF(path)
		source = "pdf"
	case ".docx":
		st, warnings, err = ExtractFromDOCX(path)
		source = "docx"
	case ".json":
		data, e := os.ReadFile(path)
		if e != nil {
			return nil, e
		}
		st = &FinanceStatement{}
		if e := json.Unmarshal(data, st); e != nil {
			return nil, fmt.Errorf("json inválido: %w", e)
		}
		if st.Schema == "" {
			st.Schema = "alset.finance.statement/1.0"
		}
		source = "json"
	default:
		return nil, fmt.Errorf("extensión no soportada: %s (usa .pdf .xlsx .xls .docx .json)", ext)
	}

	if err != nil {
		return &ExtractResult{Statement: st, Warnings: append(warnings, err.Error()), Source: source}, err
	}
	if st != nil {
		st.Source.ExtractedAt = time.Now().UTC().Format(time.RFC3339)
		if len(st.Source.Files) == 0 {
			st.Source.Files = []string{path}
		}
	}
	return &ExtractResult{Statement: st, Warnings: warnings, Source: source}, nil
}

// ToCanonicalJSON serializa el statement al formato canónico.
func (r *ExtractResult) ToCanonicalJSON() ([]byte, error) {
	if r.Statement == nil {
		return nil, fmt.Errorf("sin statement")
	}
	return json.MarshalIndent(r.Statement, "", "  ")
}
