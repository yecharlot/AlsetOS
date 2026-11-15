package declarativo

import (
	"fmt"
	"time"

	"github.com/yecharlot/AlsetOS/zyrion"
)

// RulesFor devuelve una copia de las reglas declaradas por un organismo.
// La copia evita que el consumidor modifique el ensamblaje activo.
func (rt *Runtime) RulesFor(oid string) ([]Rule, error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.Assembly == nil {
		return nil, fmt.Errorf("sin ensamblaje")
	}
	d, ok := rt.Assembly.Orgs[oid]
	if !ok {
		return nil, fmt.Errorf("organismo desconocido: %s", oid)
	}
	out := make([]Rule, len(d.Rules))
	for i, rule := range d.Rules {
		out[i] = copiarRegla(rule)
	}
	return out, nil
}

// EvaluateRuleReadOnly evalúa una regla sin escribir átomos, dependencias,
// historial ni auditoría. En el contexto MRA busca primero el átomo local
// "<oid>.<path>" y, si no existe, utiliza el path global legado.
// No ejecuta ThenPath ni concede capacidades.
func (rt *Runtime) EvaluateRuleReadOnly(oid, ruleID string) (InferenceRecord, Rule, error) {
	rules, err := rt.RulesFor(oid)
	if err != nil {
		return InferenceRecord{}, Rule{}, err
	}
	var selected *Rule
	for i := range rules {
		if rules[i].ID == ruleID {
			if selected != nil {
				return InferenceRecord{}, Rule{}, fmt.Errorf("regla duplicada %q en %s", ruleID, oid)
			}
			selected = &rules[i]
		}
	}
	if selected == nil {
		return InferenceRecord{}, Rule{}, fmt.Errorf("regla %q no declarada en %s", ruleID, oid)
	}

	rule := *selected
	evidence := make([]string, 0, len(rule.When))
	overall := zyrion.Si
	for _, condition := range rule.When {
		local := condition
		local.Path = oid + "." + condition.Path
		if _, exists := rt.State.Get(local.Path); !exists {
			if _, globalExists := rt.State.Get(condition.Path); globalExists {
				local = condition
			}
		}
		state, detail := evalCondition(rt.State, local)
		evidence = append(evidence, detail+" → "+state.Texto())
		if state == zyrion.No {
			overall = zyrion.No
		} else if state == zyrion.Incierto && overall != zyrion.No {
			overall = zyrion.Incierto
		}
	}
	record := InferenceRecord{
		TS: time.Now().UTC(), OrgOID: oid, RuleID: rule.ID,
		Zyrion: overall.Texto(), Evidence: evidence,
		ThenPath: rule.ThenPath, Skipped: true,
		Reason: "inspección de solo lectura; no se aplicó la conclusión",
	}
	return record, copiarRegla(rule), nil
}

func copiarRegla(rule Rule) Rule {
	out := rule
	out.When = append([]Condition(nil), rule.When...)
	return out
}
