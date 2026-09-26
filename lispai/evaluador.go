package lispai

import (
	"fmt"
	"strings"

	"github.com/yecharlot/AlsetOS/organismo"
)

type Resultado struct {
	Valor   string
	Cambios int
}

func (Interprete) Ejecutar(programa string, e *organismo.Organismo) (Resultado, error) {
	tokens := strings.Fields(strings.ReplaceAll(strings.ReplaceAll(programa, "(", " ( "), ")", " ) "))
	raiz, resto, err := parsear(tokens)
	if err != nil {
		return Resultado{}, err
	}
	if len(resto) > 0 {
		return Resultado{}, fmt.Errorf("tokens sobrantes")
	}
	valor, cambios, err := evaluarNodo(raiz, e)
	return Resultado{Valor: valor, Cambios: cambios}, err
}

func parsear(tokens []string) (any, []string, error) {
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("LispAI vacío")
	}
	t := tokens[0]
	if t == "(" {
		var lista []any
		tokens = tokens[1:]
		for len(tokens) > 0 && tokens[0] != ")" {
			n, resto, err := parsear(tokens)
			if err != nil {
				return nil, nil, err
			}
			lista = append(lista, n)
			tokens = resto
		}
		if len(tokens) == 0 {
			return nil, nil, fmt.Errorf("paréntesis sin cerrar")
		}
		return lista, tokens[1:], nil
	}
	if t == ")" {
		return nil, nil, fmt.Errorf("paréntesis inesperado")
	}
	return strings.Trim(t, "\""), tokens[1:], nil
}

func evaluarNodo(n any, e *organismo.Organismo) (string, int, error) {
	lista, ok := n.([]any)
	if !ok {
		return fmt.Sprint(n), 0, nil
	}
	if len(lista) == 0 {
		return "", 0, nil
	}
	op := fmt.Sprint(lista[0])
	switch op {
	case "recordar":
		if len(lista) != 3 {
			return "", 0, fmt.Errorf("recordar requiere clave y valor")
		}
		clave := fmt.Sprint(lista[1])
		valor, _, err := evaluarNodo(lista[2], e)
		if err != nil {
			return "", 0, err
		}
		e.Memoria[clave] = valor
		return valor, 1, nil
	case "borrar":
		if len(lista) != 2 {
			return "", 0, fmt.Errorf("borrar requiere clave")
		}
		delete(e.Memoria, fmt.Sprint(lista[1]))
		return "", 1, nil
	case "leer":
		if len(lista) != 2 {
			return "", 0, fmt.Errorf("leer requiere clave")
		}
		return e.Memoria[fmt.Sprint(lista[1])], 0, nil
	case "igual":
		if len(lista) != 3 {
			return "", 0, fmt.Errorf("igual requiere dos valores")
		}
		a, _, err := evaluarNodo(lista[1], e)
		if err != nil {
			return "", 0, err
		}
		b, _, err := evaluarNodo(lista[2], e)
		if err != nil {
			return "", 0, err
		}
		if a == b {
			return "si", 0, nil
		}
		return "no", 0, nil
	case "si":
		if len(lista) != 3 {
			return "", 0, fmt.Errorf("si requiere condición y acción")
		}
		cond, _, err := evaluarNodo(lista[1], e)
		if err != nil {
			return "", 0, err
		}
		if cond == "si" {
			return evaluarNodo(lista[2], e)
		}
		return "no", 0, nil
	case "secuencia":
		var valor string
		cambios := 0
		for _, parte := range lista[1:] {
			v, c, err := evaluarNodo(parte, e)
			if err != nil {
				return "", cambios, err
			}
			valor = v
			cambios += c
		}
		return valor, cambios, nil
	default:
		if len(lista) == 1 {
			return op, 0, nil
		}
		return "", 0, fmt.Errorf("forma LispAI desconocida: %s", op)
	}
}
