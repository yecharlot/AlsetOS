package ingest

import (
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

func baseName(path string) string {
	return filepath.Base(path)
}

func parseNumber(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimPrefix(s, "$")
	neg := false
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		neg = true
		s = s[1 : len(s)-1]
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		var b strings.Builder
		for _, r := range s {
			if unicode.IsDigit(r) || r == '.' || r == '-' {
				b.WriteRune(r)
			}
		}
		f, _ = strconv.ParseFloat(b.String(), 64)
	}
	if neg {
		f = -f
	}
	return f
}

func isZeroLike(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" || s == "0.0" || s == "-" {
		return true
	}
	return parseNumber(s) == 0
}
