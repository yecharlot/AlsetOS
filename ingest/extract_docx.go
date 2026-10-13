package ingest

import (
	"archive/zip"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var reWT = regexp.MustCompile(`(?s)<w:t[^>]*>(.*?)</w:t>`)

// ExtractFromDOCX extrae texto de un .docx (Office Open XML) y mapea etiquetas financieras.
// No requiere dependencias externas.
func ExtractFromDOCX(path string) (*FinanceStatement, []string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, nil, fmt.Errorf("abrir docx: %w", err)
	}
	defer r.Close()

	var text strings.Builder
	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, nil, err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, nil, err
		}
		matches := reWT.FindAllSubmatch(data, -1)
		for _, m := range matches {
			text.Write(m[1])
			text.WriteByte(' ')
		}
		break
	}

	raw := text.String()
	if len(strings.TrimSpace(raw)) < 30 {
		return &FinanceStatement{Schema: "alset.finance.statement/1.0"},
			[]string{"docx sin texto útil"}, nil
	}

	st, warnings := mapTextToStatement(raw)
	st.Source.Files = []string{path}
	st.Entity.Name = strings.TrimSuffix(baseName(path), ".docx")
	return st, warnings, nil
}
