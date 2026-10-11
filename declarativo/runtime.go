package declarativo

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/organismo"
)

// AuditEntry registro de interacciones y transiciones.
type AuditEntry struct {
	TS      time.Time `json:"ts"`
	Kind    string    `json:"kind"`
	Actor   string    `json:"actor,omitempty"`
	Detail  string    `json:"detail"`
	Payload string    `json:"payload,omitempty"`
}

// SignedPulse mensaje Pulse firmado Ed25519 (no HMAC).
type SignedPulse struct {
	Tipo      string `json:"tipo"`
	Origen    string `json:"origen"`
	Destino   string `json:"destino,omitempty"`
	Contenido string `json:"contenido"`
	Fecha     int64  `json:"fecha"`
	Nonce     string `json:"nonce"`
	PubKey    string `json:"pubkey"`
	Firma     string `json:"firma"`
}

// Runtime ensambla ORGES, mantiene estado reactivo, autoriza por capacidades.
type Runtime struct {
	mu        sync.Mutex
	Assembly  *Assembly
	State     *ReactiveState
	Instances map[string]*organismo.Organismo // OID → instancia ciclo de vida
	Caps      map[string]map[string]bool      // OID → cap set (copia autorizada)
	RootCIDs  map[string]string
	Audit     []AuditEntry
	seenNonce map[string]bool
	nodeID    *identidad.Identidad
	DataDir   string
}

func NewRuntime(dataDir string) (*Runtime, error) {
	id, err := identidad.Nueva()
	if err != nil {
		return nil, err
	}
	if dataDir != "" {
		_ = os.MkdirAll(dataDir, 0o700)
	}
	return &Runtime{
		State:     NewReactive(),
		Instances: map[string]*organismo.Organismo{},
		Caps:      map[string]map[string]bool{},
		RootCIDs:  map[string]string{},
		seenNonce: map[string]bool{},
		nodeID:    id,
		DataDir:   dataDir,
	}, nil
}

func (rt *Runtime) audit(kind, actor, detail, payload string) {
	rt.Audit = append(rt.Audit, AuditEntry{
		TS: time.Now().UTC(), Kind: kind, Actor: actor, Detail: detail, Payload: payload,
	})
}

// LoadAssembly valida, resuelve contratos e instancia organismos.
func (rt *Runtime) LoadAssembly(decls []OrgDecl) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	asm, err := Resolve(decls)
	if err != nil {
		return err
	}
	rt.Assembly = asm
	rt.Instances = map[string]*organismo.Organismo{}
	rt.Caps = map[string]map[string]bool{}
	rt.RootCIDs = map[string]string{}

	for oid, d := range asm.Orgs {
		cid, err := d.DefinitionCID()
		if err != nil {
			return err
		}
		rt.RootCIDs[oid] = cid
		o := organismo.Nuevo(d.Name)
		if err := o.Preparar(cid); err != nil {
			return fmt.Errorf("%s preparar: %w", oid, err)
		}
		// Definición canónica = DefinitionCID de la declaración de reglas (fuente de verdad).
		if err := o.VerificarDefinicion(cid, true); err != nil {
			return fmt.Errorf("%s verificar definición: %w", oid, err)
		}
		caps := map[string]bool{}
		for _, c := range d.Capabilities {
			o.OtorgarCapacidad(c)
			caps[c] = true
		}
		if err := o.Iniciar(); err != nil {
			return fmt.Errorf("%s iniciar: %w", oid, err)
		}
		rt.Instances[oid] = o
		rt.Caps[oid] = caps
		for k, v := range d.InitialState {
			rt.State.Set(k, v)
		}
		rt.audit("load", oid, "organismo ensamblado", cid)
	}
	for _, b := range asm.Bindings {
		rt.audit("bind", b.ConsumerOID, b.Reason, b.ProviderOID+"|"+b.Contract)
	}
	return nil
}

// HasCap verificación de capacidad (identidad ≠ autoridad).
func (rt *Runtime) HasCap(oid, cap string) bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	m := rt.Caps[oid]
	return m != nil && m[cap]
}

// RequireCap error si falta capacidad.
func (rt *Runtime) RequireCap(oid, cap string) error {
	if !rt.HasCap(oid, cap) {
		rt.mu.Lock()
		rt.audit("auth_deny", oid, "capacidad no concedida: "+cap, "")
		rt.mu.Unlock()
		return fmt.Errorf("capacidad no concedida: %s para %s", cap, oid)
	}
	return nil
}

// EmitPulse firma y registra un Pulse; protege replay por nonce.
func (rt *Runtime) EmitPulse(tipo, origen, destino, contenido, nonce string) (*SignedPulse, error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if nonce == "" {
		return nil, fmt.Errorf("pulse requiere nonce")
	}
	if rt.seenNonce[nonce] {
		return nil, fmt.Errorf("pulse rechazado: nonce reutilizado")
	}
	rt.seenNonce[nonce] = true
	p := &SignedPulse{
		Tipo: tipo, Origen: origen, Destino: destino, Contenido: contenido,
		Fecha: time.Now().UTC().Unix(), Nonce: nonce,
		PubKey: base64.StdEncoding.EncodeToString(rt.nodeID.ClavePublica),
	}
	payload, _ := json.Marshal(struct {
		Tipo, Origen, Destino, Contenido, Nonce string
		Fecha                                   int64
	}{p.Tipo, p.Origen, p.Destino, p.Contenido, p.Nonce, p.Fecha})
	sig := rt.nodeID.Firmar(payload)
	p.Firma = base64.StdEncoding.EncodeToString(sig)
	rt.audit("pulse", origen, tipo+":"+contenido, destino)
	return p, nil
}

// VerifyPulse comprueba firma Ed25519.
func VerifyPulse(p *SignedPulse) bool {
	if p == nil {
		return false
	}
	pub, err := base64.StdEncoding.DecodeString(p.PubKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return false
	}
	sig, err := base64.StdEncoding.DecodeString(p.Firma)
	if err != nil {
		return false
	}
	payload, _ := json.Marshal(struct {
		Tipo, Origen, Destino, Contenido, Nonce string
		Fecha                                   int64
	}{p.Tipo, p.Origen, p.Destino, p.Contenido, p.Nonce, p.Fecha})
	return identidad.Verificar(pub, payload, sig)
}

// EvalOrgRules evalúa todas las reglas ARL de un organismo.
func (rt *Runtime) EvalOrgRules(oid string) ([]InferenceRecord, error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.Assembly == nil {
		return nil, fmt.Errorf("sin ensamblaje")
	}
	d, ok := rt.Assembly.Orgs[oid]
	if !ok {
		return nil, fmt.Errorf("organismo desconocido: %s", oid)
	}
	var recs []InferenceRecord
	for _, r := range d.Rules {
		recs = append(recs, rt.State.EvalRule(oid, r))
	}
	return recs, nil
}

// EvalDirty recalcula solo átomos dirty (reactividad granular).
func (rt *Runtime) EvalDirty() []InferenceRecord {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.Assembly == nil {
		return nil
	}
	dirty := map[string]bool{}
	for _, a := range rt.State.DirtyAtoms() {
		dirty[a] = true
	}
	var recs []InferenceRecord
	for oid, d := range rt.Assembly.Orgs {
		for _, r := range d.Rules {
			if dirty[r.ThenPath] {
				recs = append(recs, rt.State.EvalRule(oid, r))
			}
		}
	}
	return recs
}

// SetState cambia un átomo (p.ej. stock) y marca dependientes.
func (rt *Runtime) SetState(path string, v interface{}) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.State.Set(path, v)
	rt.audit("state", "", path, fmt.Sprint(v))
}

func (rt *Runtime) GetState(path string) (interface{}, bool) {
	return rt.State.Get(path)
}

// Persist guarda snapshot + audit + bindings.
func (rt *Runtime) Persist() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.DataDir == "" {
		return fmt.Errorf("DataDir vacío")
	}
	doc := map[string]interface{}{
		"atoms":     rt.State.Snapshot(),
		"root_cids": rt.RootCIDs,
		"bindings":  rt.Assembly.Bindings,
		"audit":     rt.Audit,
		"history":   rt.State.HistoryCopy(),
		"caps":      rt.Caps,
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(rt.DataDir, "runtime_state.json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Restore carga estado persistido (átomos y audit); requiere LoadAssembly previo o concurrente.
func (rt *Runtime) Restore() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	path := filepath.Join(rt.DataDir, "runtime_state.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var doc struct {
		Atoms map[string]interface{} `json:"atoms"`
		Audit []AuditEntry           `json:"audit"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return err
	}
	for k, v := range doc.Atoms {
		rt.State.Atoms[k] = v
	}
	rt.Audit = append(rt.Audit, doc.Audit...)
	rt.audit("restore", "", "estado recuperado", path)
	return nil
}

// ProviderOID devuelve el provider resuelto para un contrato del consumer.
func (rt *Runtime) ProviderOID(consumerOID, contract string) (string, bool) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.Assembly == nil {
		return "", false
	}
	p, ok := rt.Assembly.ProviderFor[consumerOID+"|"+contract]
	return p, ok
}

// AuditLen número de entradas de auditoría (sin exponer estructura interna).
func (rt *Runtime) AuditLen() int {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return len(rt.Audit)
}
