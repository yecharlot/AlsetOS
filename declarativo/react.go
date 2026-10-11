package declarativo

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/yecharlot/AlsetOS/zyrion"
)

// InferenceRecord evidencia trazable de una evaluación ARL.
type InferenceRecord struct {
	TS        time.Time     `json:"ts"`
	OrgOID    string        `json:"org_oid"`
	RuleID    string        `json:"rule_id"`
	Zyrion    string        `json:"zyrion"`
	Evidence  []string      `json:"evidence"`
	ThenPath  string        `json:"then_path,omitempty"`
	ThenValue interface{}   `json:"then_value,omitempty"`
	Skipped   bool          `json:"skipped,omitempty"`
	Reason    string        `json:"reason,omitempty"`
}

// ReactiveState átomos + dependencias explícitas + historial (AlsetState extendido).
type ReactiveState struct {
	mu      sync.RWMutex
	Atoms   map[string]interface{}
	// Deps[atom] = paths de los que depende (para invalidación granular)
	Deps    map[string][]string
	Dirty   map[string]bool
	History []InferenceRecord
}

func NewReactive() *ReactiveState {
	return &ReactiveState{
		Atoms: map[string]interface{}{},
		Deps:  map[string][]string{},
		Dirty: map[string]bool{},
	}
}

func (s *ReactiveState) Get(path string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Atoms[path]
	return v, ok
}

func (s *ReactiveState) Set(path string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Atoms[path] = v
	// marcar dependientes como dirty
	for atom, deps := range s.Deps {
		for _, d := range deps {
			if d == path {
				s.Dirty[atom] = true
			}
		}
	}
}

func (s *ReactiveState) RegisterDeps(atom string, deps []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := append([]string{}, deps...)
	sort.Strings(cp)
	s.Deps[atom] = cp
	s.Dirty[atom] = true
}

func (s *ReactiveState) Snapshot() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := map[string]interface{}{}
	for k, v := range s.Atoms {
		out[k] = v
	}
	return out
}

func (s *ReactiveState) HistoryCopy() []InferenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]InferenceRecord, len(s.History))
	copy(out, s.History)
	return out
}

func asFloat(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case jsonNumber:
		return float64(t), true
	default:
		return 0, false
	}
}

// jsonNumber avoids importing encoding/json number type issues
type jsonNumber float64

func evalCondition(s *ReactiveState, c Condition) (zyrion.Valor, string) {
	v, ok := s.Get(c.Path)
	ev := fmt.Sprintf("%s %s %v (got=%v present=%v)", c.Path, c.Op, c.Value, v, ok)
	switch c.Op {
	case "exists":
		if ok {
			return zyrion.Si, ev
		}
		return zyrion.No, ev
	case "missing":
		if !ok {
			return zyrion.Si, ev
		}
		return zyrion.No, ev
	}
	if !ok {
		return zyrion.Incierto, ev + " [dato ausente→incierto]"
	}
	switch c.Op {
	case "eq":
		if fmt.Sprint(v) == fmt.Sprint(c.Value) {
			return zyrion.Si, ev
		}
		return zyrion.No, ev
	case "ne":
		if fmt.Sprint(v) != fmt.Sprint(c.Value) {
			return zyrion.Si, ev
		}
		return zyrion.No, ev
	case "gt", "gte", "lt", "lte":
		lf, ok1 := asFloat(v)
		rf, ok2 := asFloat(c.Value)
		if !ok1 || !ok2 {
			return zyrion.Incierto, ev + " [no numérico]"
		}
		var pass bool
		switch c.Op {
		case "gt":
			pass = lf > rf
		case "gte":
			pass = lf >= rf
		case "lt":
			pass = lf < rf
		case "lte":
			pass = lf <= rf
		}
		if pass {
			return zyrion.Si, ev
		}
		return zyrion.No, ev
	default:
		return zyrion.Incierto, ev + " [op desconocida]"
	}
}

// EvalRule evalúa una regla ARL con Zyrion (AND de condiciones).
// No concede capacidades: solo escribe estado si procede.
func (s *ReactiveState) EvalRule(orgOID string, r Rule) InferenceRecord {
	deps := make([]string, 0, len(r.When))
	for _, w := range r.When {
		deps = append(deps, w.Path)
	}
	s.RegisterDeps(r.ThenPath, deps)

	evidence := []string{}
	overall := zyrion.Si
	for _, w := range r.When {
		z, ev := evalCondition(s, w)
		evidence = append(evidence, ev+" → "+z.Texto())
		if z == zyrion.No {
			overall = zyrion.No
		} else if z == zyrion.Incierto && overall != zyrion.No {
			overall = zyrion.Incierto
		}
	}

	rec := InferenceRecord{
		TS: time.Now().UTC(), OrgOID: orgOID, RuleID: r.ID,
		Zyrion: overall.Texto(), Evidence: evidence, ThenPath: r.ThenPath,
	}

	switch overall {
	case zyrion.Si:
		s.Set(r.ThenPath, r.ThenValue)
		rec.ThenValue = r.ThenValue
	case zyrion.Incierto:
		if r.ThenOnIndeterminate != nil {
			s.Set(r.ThenPath, r.ThenOnIndeterminate)
			rec.ThenValue = r.ThenOnIndeterminate
			rec.Reason = "condiciones indeterminadas; valor de precaución"
		} else {
			rec.Skipped = true
			rec.Reason = "no se finge certeza (Zyrion=incierto)"
		}
	default:
		// Conclusión negativa explícita (no conservar valor previo de éxito).
		neg := interface{}("no")
		if _, isBool := r.ThenValue.(bool); isBool {
			neg = false
		}
		s.Set(r.ThenPath, neg)
		rec.ThenValue = neg
		rec.Reason = "condiciones no satisfechas (Zyrion=no)"
	}

	s.mu.Lock()
	s.History = append(s.History, rec)
	if len(s.History) > 500 {
		s.History = s.History[len(s.History)-500:]
	}
	delete(s.Dirty, r.ThenPath)
	s.mu.Unlock()
	return rec
}

// DirtyAtoms lista conclusiones a recalcular.
func (s *ReactiveState) DirtyAtoms() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []string
	for k, d := range s.Dirty {
		if d {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
