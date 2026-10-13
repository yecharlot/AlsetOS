package ingest

import (
	"fmt"
)

// ExtractFromXLS: en esta rebanada el camino principal es JSON.
// Para XLSX: go get github.com/xuri/excelize/v2 y restaura la implementación
// documentada en docs/EXTRACTORES_PDF_XLS.md.
func ExtractFromXLS(path string) (*FinanceStatement, []string, error) {
	return nil, []string{
		"XLS/XLSX: usa JSON canónico (ejemplos/finance/sample.json) o instala excelize",
	}, fmt.Errorf("extractor XLS no enlazado; use .json (%s)", path)
}
