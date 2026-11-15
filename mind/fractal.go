package mind

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/yecharlot/AlsetOS/declarativo"
	"github.com/yecharlot/AlsetOS/zyrion"
)

const (
	// TipoRelacionFractal habilita explícitamente una arista para expandir
	// una evaluación indeterminada al siguiente nivel.
	TipoRelacionFractal              = "contexto-fractal"
	profundidadFractalPredeterminada = 8
	maxOrganismosFractal             = 256
)

// NivelEvaluacion conserva el resultado y la procedencia de un nivel.
// Hijos contiene las evaluaciones del mismo contrato de regla en organismos
// conectados mediante relaciones explícitamente fractales.
type NivelEvaluacion struct {
	OID         string            `json:"oid"`
	Profundidad int               `json:"profundidad"`
	Estado      zyrion.Valor      `json:"estado"`
	Evidencias  []string          `json:"evidencias,omitempty"`
	Motivo      string            `json:"motivo,omitempty"`
	Hijos       []NivelEvaluacion `json:"hijos,omitempty"`
}

// EvaluacionFractal es una traza completa. No es una autorización ni ejecuta
// acciones: representa exclusivamente el resultado lógico y su procedencia.
type EvaluacionFractal struct {
	ReglaID             string          `json:"regla_id"`
	Estado              zyrion.Valor    `json:"estado"`
	Raiz                NivelEvaluacion `json:"raiz"`
	ProfundidadMaxima   int             `json:"profundidad_maxima"`
	OrganismosEvaluados int             `json:"organismos_evaluados"`
}

// EvaluarFractal aplica una misma regla declarada a través de niveles de
// organismos relacionados cuando Zyrion devuelve 2 (indeterminado).
//
// Solo las relaciones cuyo Tipo es TipoRelacionFractal habilitan la expansión.
// Política de resolución: un nivel solo se resuelve si todos sus hijos
// aplicables terminan en el mismo estado determinado (0 o 1). Si hay
// desacuerdo, incertidumbre, ciclo o límite alcanzado, permanece en 2.
// Las reglas se comparan estructuralmente: no se sustituye una regla por
// otra distinta solo porque comparta un ID.
func (Mente) EvaluarFractal(rt *declarativo.Runtime, oid, ruleID string, maxDepth int) (EvaluacionFractal, error) {
	if rt == nil {
		return EvaluacionFractal{}, fmt.Errorf("runtime nulo")
	}
	if oid == "" || ruleID == "" {
		return EvaluacionFractal{}, fmt.Errorf("oid y ruleID son obligatorios")
	}
	if maxDepth <= 0 {
		maxDepth = profundidadFractalPredeterminada
	}
	rootRules, err := rt.RulesFor(oid)
	if err != nil {
		return EvaluacionFractal{}, err
	}
	var rootRule *declarativo.Rule
	for i := range rootRules {
		if rootRules[i].ID == ruleID {
			if rootRule != nil {
				return EvaluacionFractal{}, fmt.Errorf("regla duplicada %q en %s", ruleID, oid)
			}
			r := rootRules[i]
			rootRule = &r
		}
	}
	if rootRule == nil {
		return EvaluacionFractal{}, fmt.Errorf("regla %q no declarada en %s", ruleID, oid)
	}

	evaluated := 0
	var walk func(string, int, map[string]bool) (NivelEvaluacion, error)
	walk = func(current string, depth int, path map[string]bool) (NivelEvaluacion, error) {
		if evaluated >= maxOrganismosFractal {
			return NivelEvaluacion{OID: current, Profundidad: depth, Estado: zyrion.Incierto, Motivo: "límite de organismos alcanzado"}, nil
		}
		evaluated++
		record, rule, err := rt.EvaluateRuleReadOnly(current, ruleID)
		if err != nil {
			return NivelEvaluacion{}, err
		}
		if !reflect.DeepEqual(rule, *rootRule) {
			return NivelEvaluacion{}, fmt.Errorf("la regla %q en %s no conserva la misma definición", ruleID, current)
		}
		node := NivelEvaluacion{
			OID: current, Profundidad: depth,
			Estado: estadoDesdeTexto(record.Zyrion),
			Evidencias: append([]string(nil), record.Evidence...),
			Motivo: record.Reason,
		}
		if node.Estado != zyrion.Incierto {
			node.Motivo = "resultado determinado en este nivel"
			return node, nil
		}
		if depth >= maxDepth {
			node.Motivo = "estado 2: límite de profundidad alcanzado"
			return node, nil
		}
		neighbors, err := vecinosFractales(rt, current)
		if err != nil {
			return NivelEvaluacion{}, err
		}
		for _, neighbor := range neighbors {
			if path[neighbor] {
				node.Hijos = append(node.Hijos, NivelEvaluacion{
					OID: neighbor, Profundidad: depth + 1,
					Estado: zyrion.Incierto, Motivo: "ciclo relacional detectado; rama no reevaluada",
				})
				continue
			}
			childRules, err := rt.RulesFor(neighbor)
			if err != nil {
				return NivelEvaluacion{}, err
			}
			matching := false
			for _, candidate := range childRules {
				if candidate.ID == ruleID && reflect.DeepEqual(candidate, *rootRule) {
					matching = true
					break
				}
			}
			if !matching {
				continue
			}
			nextPath := make(map[string]bool, len(path)+1)
			for id, inPath := range path {
				nextPath[id] = inPath
			}
			nextPath[current] = true
			child, err := walk(neighbor, depth+1, nextPath)
			if err != nil {
				return NivelEvaluacion{}, err
			}
			node.Hijos = append(node.Hijos, child)
		}
		node.Estado = consensoHijos(node.Hijos)
		switch node.Estado {
		case zyrion.Si:
			node.Motivo = "todos los niveles hijos aplicables convergen en True"
		case zyrion.No:
			node.Motivo = "todos los niveles hijos aplicables convergen en False"
		default:
			if len(node.Hijos) == 0 {
				node.Motivo = "estado 2: no existe un nivel fractal relacionado con la misma regla"
			} else {
				node.Motivo = "estado 2: los niveles hijos no alcanzan un consenso determinado"
			}
		}
		return node, nil
	}

	root, err := walk(oid, 0, map[string]bool{})
	if err != nil {
		return EvaluacionFractal{}, err
	}
	result := EvaluacionFractal{
		ReglaID: ruleID, Estado: root.Estado, Raiz: root,
		ProfundidadMaxima: maxDepth, OrganismosEvaluados: evaluated,
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return EvaluacionFractal{}, fmt.Errorf("serializar traza fractal: %w", err)
	}
	if err := rt.RecordEvaluation("mind.fractal.v1", oid, ruleID, int(result.Estado), payload); err != nil {
		return EvaluacionFractal{}, fmt.Errorf("registrar traza fractal: %w", err)
	}
	return result, nil
}

func vecinosFractales(rt *declarativo.Runtime, oid string) ([]string, error) {
	context, err := rt.RelationalContext(oid)
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for _, edge := range context.Salientes {
		if edge.Tipo == TipoRelacionFractal {
			set[edge.Destino] = true
		}
	}
	for _, edge := range context.Entrantes {
		if edge.Tipo == TipoRelacionFractal {
			set[edge.Origen] = true
		}
	}
	out := make([]string, 0, len(set))
	for neighbor := range set {
		out = append(out, neighbor)
	}
	sort.Strings(out)
	return out, nil
}

func estadoDesdeTexto(value string) zyrion.Valor {
	switch value {
	case "si":
		return zyrion.Si
	case "no":
		return zyrion.No
	default:
		return zyrion.Incierto
	}
}

// consensoHijos solo resuelve cuando hay al menos un hijo y todos los hijos
// son determinados y coinciden. No convierte ausencia o conflicto en True/False.
func consensoHijos(children []NivelEvaluacion) zyrion.Valor {
	if len(children) == 0 {
		return zyrion.Incierto
	}
	result := children[0].Estado
	if result == zyrion.Incierto {
		return zyrion.Incierto
	}
	for _, child := range children[1:] {
		if child.Estado != result {
			return zyrion.Incierto
		}
	}
	return result
}
