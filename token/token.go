// Package token implementa tokenización de modelos de negocio al estilo Alset/PrismaTEC:
// manifiesto .aos versionado + huella content-addressed + cadena append-only de eventos.
// No imita blockchains de terceros: es evidencia soberana del organismo de negocio.
package token

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// BusinessToken es la unidad tokenizable: un proceso, nicho, evento o modelo completo.
type BusinessToken struct {
	ID          string            `json:"id"`
	Kind        string            `json:"kind"` // process | niche | event | model | flow
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	ManifestCID string            `json:"manifest_cid"` // cid1:sha256 del .aos
	Version     string            `json:"version"`
	Domain      string            `json:"domain,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
	CreatedAt   string            `json:"created_at"`
	ChainHead   string            `json:"chain_head"` // hash del último eslabón
}

// ChainEvent es un eslabón de evidencia (análogo a Pulse + hash chain de PrismaTEC).
type ChainEvent struct {
	Seq       uint64         `json:"seq"`
	PrevHash  string         `json:"prev_hash"`
	Hash      string         `json:"hash"`
	Type      string         `json:"type"` // mint | mutate | test | deploy | audit | pulse
	Target    string         `json:"target"`
	Payload   map[string]any `json:"payload,omitempty"`
	Timestamp string         `json:"timestamp"`
}

// Registry almacena tokens y cadenas en disco (JSON), portable y offline-first.
type Registry struct {
	mu     sync.Mutex
	dir    string
	tokens map[string]*BusinessToken
	chains map[string][]ChainEvent
}

func NewRegistry(dir string) (*Registry, error) {
	if dir == "" {
		dir = "data/tokens"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	r := &Registry{dir: dir, tokens: map[string]*BusinessToken{}, chains: map[string][]ChainEvent{}}
	_ = r.load()
	return r, nil
}

// FingerprintAOS genera CID estilo PrismaTEC: "cid1:" + hex(sha256(canonical)).
func FingerprintAOS(source string) string {
	sum := sha256.Sum256([]byte(source))
	return "cid1:" + hex.EncodeToString(sum[:])
}

func hashEvent(prev string, seq uint64, typ, target, ts string, payload map[string]any) string {
	b, _ := json.Marshal(struct {
		Prev    string         `json:"prev"`
		Seq     uint64         `json:"seq"`
		Type    string         `json:"type"`
		Target  string         `json:"target"`
		TS      string         `json:"ts"`
		Payload map[string]any `json:"payload"`
	}{prev, seq, typ, target, ts, payload})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Mint tokeniza un manifiesto .aos (proceso/nicho/evento/modelo).
func (r *Registry) Mint(kind, name, desc, version, domain, aos string, meta map[string]string) (*BusinessToken, ChainEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if version == "" {
		version = "1.0.0"
	}
	if kind == "" {
		kind = "model"
	}
	cid := FingerprintAOS(aos)
	id := fmt.Sprintf("%s-%s", sanitize(name), cid[len(cid)-12:])
	ts := time.Now().UTC().Format(time.RFC3339)
	ev := ChainEvent{
		Seq: 1, PrevHash: "", Type: "mint", Target: id,
		Payload: map[string]any{
			"kind": kind, "name": name, "manifest_cid": cid, "version": version, "domain": domain,
		},
		Timestamp: ts,
	}
	ev.Hash = hashEvent(ev.PrevHash, ev.Seq, ev.Type, ev.Target, ev.Timestamp, ev.Payload)
	tok := &BusinessToken{
		ID: id, Kind: kind, Name: name, Description: desc,
		ManifestCID: cid, Version: version, Domain: domain, Meta: meta,
		CreatedAt: ts, ChainHead: ev.Hash,
	}
	r.tokens[id] = tok
	r.chains[id] = []ChainEvent{ev}
	// guardar manifiesto junto al token
	_ = os.WriteFile(filepath.Join(r.dir, id+".aos"), []byte(aos), 0o644)
	if err := r.persist(); err != nil {
		return nil, ev, err
	}
	return tok, ev, nil
}

// Append añade evidencia (deploy, test, audit, pulse de negocio).
func (r *Registry) Append(tokenID, typ string, payload map[string]any) (ChainEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tok, ok := r.tokens[tokenID]
	if !ok {
		return ChainEvent{}, fmt.Errorf("token no encontrado: %s", tokenID)
	}
	chain := r.chains[tokenID]
	prev := ""
	seq := uint64(1)
	if len(chain) > 0 {
		prev = chain[len(chain)-1].Hash
		seq = chain[len(chain)-1].Seq + 1
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	ev := ChainEvent{Seq: seq, PrevHash: prev, Type: typ, Target: tokenID, Payload: payload, Timestamp: ts}
	ev.Hash = hashEvent(prev, seq, typ, tokenID, ts, payload)
	r.chains[tokenID] = append(chain, ev)
	tok.ChainHead = ev.Hash
	return ev, r.persist()
}

func (r *Registry) Get(id string) (*BusinessToken, []ChainEvent, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tokens[id]
	if !ok {
		return nil, nil, false
	}
	ch := append([]ChainEvent(nil), r.chains[id]...)
	cp := *t
	return &cp, ch, true
}

func (r *Registry) List() []*BusinessToken {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*BusinessToken, 0, len(r.tokens))
	for _, t := range r.tokens {
		cp := *t
		out = append(out, &cp)
	}
	return out
}

func (r *Registry) persist() error {
	type disk struct {
		Tokens map[string]*BusinessToken `json:"tokens"`
		Chains map[string][]ChainEvent   `json:"chains"`
	}
	d := disk{Tokens: r.tokens, Chains: r.chains}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(r.dir, "registry.json"), b, 0o644)
}

func (r *Registry) load() error {
	p := filepath.Join(r.dir, "registry.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	var d struct {
		Tokens map[string]*BusinessToken `json:"tokens"`
		Chains map[string][]ChainEvent   `json:"chains"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		return err
	}
	if d.Tokens != nil {
		r.tokens = d.Tokens
	}
	if d.Chains != nil {
		r.chains = d.Chains
	}
	return nil
}

func sanitize(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
			out = append(out, c)
		} else if c == ' ' || c == '_' {
			out = append(out, '-')
		}
	}
	if len(out) == 0 {
		return "negocio"
	}
	if len(out) > 40 {
		out = out[:40]
	}
	return string(out)
}
