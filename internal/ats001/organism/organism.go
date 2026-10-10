package organism

import (
	"github.com/yecharlot/AlsetOS/internal/ats001/identity"
	"github.com/yecharlot/AlsetOS/internal/ats001/mind"
	"github.com/yecharlot/AlsetOS/internal/ats001/state"
)

type Cap struct {
	ID       string   `json:"id"`
	Emisor   string   `json:"emisor"`
	Receptor string   `json:"receptor"`
	Accion   string   `json:"accion"`
	Version  string   `json:"version"`
}

type Organism struct {
	Identity identity.RootIdentity `json:"identity"`
	State    *state.AlsetState     `json:"-"`
	Mind     *mind.Mind            `json:"mind"`
	Caps     []Cap                 `json:"capabilities"`
	Phase    string                `json:"phase"`
	Master   bool                  `json:"master"`
	Mailbox  []map[string]interface{} `json:"mailbox"`
	Kind     string                `json:"kind"` // orges kind
}

func New(id identity.RootIdentity, kind string) *Organism {
	return &Organism{
		Identity: id,
		State:    state.New(),
		Mind:     mind.New(),
		Phase:    "operacion",
		Kind:     kind,
	}
}

func (o *Organism) OID() string     { return o.Identity.OID }
func (o *Organism) RootCID() string { return o.Identity.RootCID }

func (o *Organism) HasCap(id string) bool {
	if o.Master {
		return true
	}
	for _, c := range o.Caps {
		if c.ID == id {
			return true
		}
	}
	return false
}

func (o *Organism) Public() map[string]interface{} {
	caps := []string{}
	for _, c := range o.Caps {
		caps = append(caps, c.ID)
	}
	goals := []string{}
	for _, g := range o.Mind.Goals {
		goals = append(goals, g.Name)
	}
	return map[string]interface{}{
		"oid":          o.OID(),
		"root_cid":     o.RootCID(),
		"phase":        o.Phase,
		"master":       o.Master,
		"kind":         o.Kind,
		"capabilities": caps,
		"goals":        goals,
		"mailbox":      len(o.Mailbox),
		"state":        o.State.Snapshot(),
		"kpub":         o.Identity.Keys.Public,
	}
}
