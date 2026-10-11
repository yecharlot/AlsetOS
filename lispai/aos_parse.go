package lispai

import (
	"fmt"
	"strings"
	"unicode"
)

// AOSNode es un átomo (string) o lista de nodos (S-expresión).
type AOSNode any

// TokenizarAOS divide un programa LISPAI-AOS en tokens.
func TokenizarAOS(src string) []string {
	var tokens []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			tokens = append(tokens, b.String())
			b.Reset()
		}
	}
	inStr := false
	for i := 0; i < len(src); i++ {
		c := rune(src[i])
		if inStr {
			if c == '"' {
				inStr = false
				b.WriteRune(c)
				flush()
			} else {
				b.WriteRune(c)
			}
			continue
		}
		switch {
		case c == ';':
			flush()
			for i+1 < len(src) && src[i+1] != '\n' {
				i++
			}
		case c == '"':
			flush()
			inStr = true
			b.WriteRune(c)
		case c == '(' || c == ')':
			flush()
			tokens = append(tokens, string(c))
		case unicode.IsSpace(c):
			flush()
		default:
			b.WriteRune(c)
		}
	}
	flush()
	return tokens
}

// ParsearAOS parsea tokens a un árbol de S-expresiones (lista de forms de top-level).
func ParsearAOS(tokens []string) ([]AOSNode, error) {
	var forms []AOSNode
	rest := tokens
	for len(rest) > 0 {
		n, r, err := parseAOSNode(rest)
		if err != nil {
			return nil, err
		}
		forms = append(forms, n)
		rest = r
	}
	return forms, nil
}

func parseAOSNode(tokens []string) (AOSNode, []string, error) {
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("AOS: entrada vacía")
	}
	if tokens[0] == "(" {
		var list []AOSNode
		rest := tokens[1:]
		for len(rest) > 0 && rest[0] != ")" {
			n, r, err := parseAOSNode(rest)
			if err != nil {
				return nil, nil, err
			}
			list = append(list, n)
			rest = r
		}
		if len(rest) == 0 {
			return nil, nil, fmt.Errorf("AOS: paréntesis sin cerrar")
		}
		return list, rest[1:], nil
	}
	if tokens[0] == ")" {
		return nil, nil, fmt.Errorf("AOS: paréntesis de cierre inesperado")
	}
	return unquote(tokens[0]), tokens[1:], nil
}

func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func asList(n AOSNode) ([]AOSNode, bool) {
	l, ok := n.([]AOSNode)
	return l, ok
}

func asAtom(n AOSNode) string {
	return fmt.Sprint(n)
}
