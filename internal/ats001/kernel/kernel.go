package kernel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yecharlot/AlsetOS/internal/ats001/identity"
	"github.com/yecharlot/AlsetOS/internal/ats001/organism"
	"github.com/yecharlot/AlsetOS/internal/ats001/pulse"
)

// Microkernel ATS-001 Parte XIII
type Kernel struct {
	mu        sync.Mutex
	DataDir   string
	MasterOID string
	Orgs      map[string]*organism.Organism
	Ident     map[string]identity.RootIdentity
	Nonces    map[string]bool
	Ticks     int
	Audit     []map[string]interface{}
}

func New(dataDir string) *Kernel {
	_ = os.MkdirAll(filepath.Join(dataDir, "organisms"), 0o755)
	return &Kernel{
		DataDir: dataDir,
		Orgs:    map[string]*organism.Organism{},
		Ident:   map[string]identity.RootIdentity{},
		Nonces:  map[string]bool{},
	}
}

func (k *Kernel) audit(ev string, d map[string]interface{}) {
	row := map[string]interface{}{"ts": time.Now().Unix(), "event": ev}
	for kk, vv := range d {
		row[kk] = vv
	}
	k.Audit = append(k.Audit, row)
	if len(k.Audit) > 500 {
		k.Audit = k.Audit[len(k.Audit)-500:]
	}
}

func (k *Kernel) BootMaster() *organism.Organism {
	k.mu.Lock()
	defer k.mu.Unlock()
	id := identity.Create("Master", map[string]interface{}{"role": "sovereign", "ats": "001"})
	o := organism.New(id, "master")
	o.Master = true
	o.Mind.AddGoal("soberania", 1)
	o.Mind.AddGoal("gobierno", 1)
	for _, c := range []string{"organism.create", "organism.destroy", "pulse.send", "pulse.receive", "orges.deploy", "node.shutdown"} {
		o.Caps = append(o.Caps, organism.Cap{ID: c, Emisor: id.OID, Receptor: "*", Accion: c, Version: "1.0.0"})
	}
	k.Orgs[o.OID()] = o
	k.Ident[o.OID()] = id
	k.MasterOID = o.OID()
	k.audit("boot", map[string]interface{}{"master": o.OID()})
	k.checkpointLocked(o.OID())
	return o
}

func (k *Kernel) require(oid, cap string) (*organism.Organism, error) {
	o, ok := k.Orgs[oid]
	if !ok {
		return nil, fmt.Errorf("unknown organism")
	}
	if !o.HasCap(cap) {
		return nil, fmt.Errorf("missing capability: %s", cap)
	}
	return o, nil
}

type CreateReq struct {
	ActorOID string                 `json:"actor_oid"`
	Name     string                 `json:"name"`
	Kind     string                 `json:"kind"`
	Desc     map[string]interface{} `json:"desc"`
	Caps     []string               `json:"caps"`
	Goals    []string               `json:"goals"`
}

func (k *Kernel) CreateOrganism(r CreateReq) (*organism.Organism, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if _, err := k.require(r.ActorOID, "organism.create"); err != nil {
		return nil, err
	}
	if r.Desc == nil {
		r.Desc = map[string]interface{}{}
	}
	r.Desc["kind"] = r.Kind
	id := identity.Create(r.Name, r.Desc)
	o := organism.New(id, r.Kind)
	for _, g := range r.Goals {
		o.Mind.AddGoal(g, 1)
	}
	caps := r.Caps
	if len(caps) == 0 {
		caps = []string{"pulse.receive"}
	}
	hasRecv := false
	for _, c := range caps {
		if c == "pulse.receive" {
			hasRecv = true
		}
		o.Caps = append(o.Caps, organism.Cap{ID: c, Emisor: r.ActorOID, Receptor: id.OID, Accion: c, Version: "1.0.0"})
	}
	if !hasRecv {
		o.Caps = append(o.Caps, organism.Cap{ID: "pulse.receive", Emisor: r.ActorOID, Receptor: id.OID, Accion: "receive", Version: "1.0.0"})
	}
	k.Orgs[o.OID()] = o
	k.Ident[o.OID()] = id
	k.audit("create", map[string]interface{}{"oid": o.OID(), "kind": r.Kind})
	k.checkpointLocked(o.OID())
	return o, nil
}

// CreateORGES — organismo especializado (Parte IX)
func (k *Kernel) CreateORGES(actor, name, kind string, goals, caps []string) (*organism.Organism, error) {
	return k.CreateOrganism(CreateReq{
		ActorOID: actor,
		Name:     name,
		Kind:     "orges." + kind,
		Desc:     map[string]interface{}{"orges": true, "domain": kind},
		Caps:     caps,
		Goals:    goals,
	})
}

func (k *Kernel) Pulse(src, dst string, payload map[string]interface{}, caps []string) (pulse.Message, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if _, err := k.require(src, "pulse.send"); err != nil {
		return pulse.Message{}, err
	}
	dest, ok := k.Orgs[dst]
	if !ok {
		return pulse.Message{}, fmt.Errorf("dest unknown")
	}
	if !dest.HasCap("pulse.receive") {
		return pulse.Message{}, fmt.Errorf("dest cannot receive")
	}
	id := k.Ident[src]
	msg := pulse.Make(id, dst, payload, caps)
	if k.Nonces[msg.Nonce] {
		return pulse.Message{}, fmt.Errorf("replay")
	}
	k.Nonces[msg.Nonce] = true
	if !pulse.Verify(id, msg) {
		return pulse.Message{}, fmt.Errorf("bad signature")
	}
	b, _ := json.Marshal(msg)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	dest.Mailbox = append(dest.Mailbox, m)
	dest.Mind.Remember(map[string]interface{}{"pulse": m}, false)
	k.audit("pulse", map[string]interface{}{"from": src, "to": dst, "cid": msg.PayloadCID})
	return msg, nil
}

func (k *Kernel) Tick() map[string]interface{} {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.Ticks++
	var pick *organism.Organism
	max := -1
	for _, o := range k.Orgs {
		if len(o.Mailbox) > max {
			max = len(o.Mailbox)
			pick = o
		}
	}
	delivered := 0
	if pick != nil {
		for len(pick.Mailbox) > 0 {
			pick.Mailbox = pick.Mailbox[1:]
			delivered++
		}
		pick.State.Set("last_tick", k.Ticks)
	}
	return map[string]interface{}{"tick": k.Ticks, "oid": pickOID(pick), "delivered": delivered}
}

func pickOID(o *organism.Organism) string {
	if o == nil {
		return ""
	}
	return o.OID()
}

func (k *Kernel) checkpointLocked(oid string) {
	o := k.Orgs[oid]
	if o == nil {
		return
	}
	path := filepath.Join(k.DataDir, "organisms", sanitize(oid)+".json")
	b, _ := json.MarshalIndent(o.Public(), "", "  ")
	_ = os.WriteFile(path, b, 0o644)
}

func sanitize(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			out = append(out, c)
		} else {
			out = append(out, '_')
		}
	}
	return string(out)
}

func (k *Kernel) Status() map[string]interface{} {
	k.mu.Lock()
	defer k.mu.Unlock()
	list := []map[string]interface{}{}
	for _, o := range k.Orgs {
		list = append(list, o.Public())
	}
	audit := k.Audit
	if len(audit) > 30 {
		audit = audit[len(audit)-30:]
	}
	return map[string]interface{}{
		"spec":       "ATS-001",
		"product":    "AlsetOS Native (Go)",
		"master_oid": k.MasterOID,
		"ticks":      k.Ticks,
		"organisms":  list,
		"audit":      audit,
	}
}
