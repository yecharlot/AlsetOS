package declarativo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

// VersionSnapshot identifica el contrato de serialización del Runtime.
const VersionSnapshot = 1

type EvaluationRecord struct {
	Schema     string          `json:"schema"`
	RootOID    string          `json:"root_oid"`
	RuleID     string          `json:"rule_id"`
	State      int             `json:"state"`
	RecordedAt time.Time       `json:"recorded_at"`
	Payload    json.RawMessage `json:"payload"`
	Digest     string          `json:"digest"`
}

type runtimeSnapshot struct {
	SchemaVersion       int               `json:"schema_version"`
	CreatedAt           time.Time         `json:"created_at"`
	AssemblyFingerprint string            `json:"assembly_fingerprint"`
	RootCIDs            map[string]string `json:"root_cids"`
	Bindings            []Binding         `json:"bindings"`
	Atoms               map[string]interface{} `json:"atoms"`
	Deps                map[string][]string `json:"deps"`
	Dirty               map[string]bool    `json:"dirty"`
	History             []InferenceRecord  `json:"history"`
	Audit               []AuditEntry       `json:"audit"`
	Caps                map[string]map[string]bool `json:"caps"`
	Evaluations         []EvaluationRecord `json:"evaluations"`
}

type snapshotEnvelope struct {
	Version   int             `json:"version"`
	Algorithm string          `json:"algorithm"`
	Checksum  string          `json:"checksum"`
	Payload   json.RawMessage `json:"payload"`
}

// RecordEvaluation conserva una evaluación como evidencia reproducible.
// El registro no cambia átomos, reglas, capacidades ni auditoría de dominio.
func (rt *Runtime) RecordEvaluation(schema, rootOID, ruleID string, state int, payload []byte) error {
	if schema == "" || rootOID == "" || ruleID == "" {
		return fmt.Errorf("schema, rootOID y ruleID son obligatorios")
	}
	if state < 0 || state > 2 {
		return fmt.Errorf("estado Zyrion inválido: %d", state)
	}
	if !json.Valid(payload) {
		return fmt.Errorf("payload de evaluación no es JSON válido")
	}
	var decoded interface{}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return fmt.Errorf("decodificar payload de evaluación: %w", err)
	}
	copyPayload, err := json.Marshal(decoded)
	if err != nil {
		return fmt.Errorf("canonicalizar payload de evaluación: %w", err)
	}
	sum := sha256.Sum256(copyPayload)
	record := EvaluationRecord{
		Schema: schema, RootOID: rootOID, RuleID: ruleID, State: state,
		RecordedAt: time.Now().UTC(), Payload: copyPayload, Digest: hex.EncodeToString(sum[:]),
	}
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.Assembly == nil {
		return fmt.Errorf("no se puede registrar evaluación sin ensamblaje")
	}
	if _, ok := rt.Assembly.Orgs[rootOID]; !ok {
		return fmt.Errorf("organismo raíz desconocido: %s", rootOID)
	}
	rt.Evaluations = append(rt.Evaluations, record)
	return nil
}

func (rt *Runtime) EvaluationsCopy() []EvaluationRecord {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	raw, err := json.Marshal(rt.Evaluations)
	if err != nil { return nil }
	var out []EvaluationRecord
	if json.Unmarshal(raw, &out) != nil { return nil }
	return out
}

// Persist escribe un snapshot versionado y protegido por checksum.
// El archivo temporal se sincroniza antes del rename para evitar publicar
// una escritura incompleta como estado confirmado.
func (rt *Runtime) Persist() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.DataDir == "" { return fmt.Errorf("DataDir vacío") }
	if rt.Assembly == nil { return fmt.Errorf("no hay ensamblaje cargado") }

	state, err := rt.State.ExportSnapshot()
	if err != nil { return err }
	payload := runtimeSnapshot{
		SchemaVersion: VersionSnapshot,
		CreatedAt: time.Now().UTC(),
		AssemblyFingerprint: huellaEnsamblaje(rt.RootCIDs),
		RootCIDs: copiarMapaString(rt.RootCIDs),
		Bindings: append([]Binding(nil), rt.Assembly.Bindings...),
		Atoms: state.Atoms, Deps: state.Deps, Dirty: state.Dirty,
		History: state.History, Audit: append([]AuditEntry(nil), rt.Audit...),
		Caps: copiarCapacidades(rt.Caps),
		Evaluations: copiarEvaluaciones(rt.Evaluations),
	}
	canonical, err := json.Marshal(payload)
	if err != nil { return fmt.Errorf("serializar snapshot: %w", err) }
	sum := sha256.Sum256(canonical)
	envelope := snapshotEnvelope{
		Version: VersionSnapshot, Algorithm: "sha256",
		Checksum: hex.EncodeToString(sum[:]), Payload: canonical,
	}
	encoded, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil { return fmt.Errorf("serializar envoltura del snapshot: %w", err) }
	if err := os.MkdirAll(rt.DataDir, 0o700); err != nil {
		return fmt.Errorf("crear directorio de datos: %w", err)
	}
	path := filepath.Join(rt.DataDir, "runtime_state.json")
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil { return fmt.Errorf("crear snapshot temporal: %w", err) }
	ok := false
	defer func() { if !ok { _ = os.Remove(tmp) } }()
	if _, err = file.Write(encoded); err != nil { _ = file.Close(); return fmt.Errorf("escribir snapshot temporal: %w", err) }
	if err = file.Sync(); err != nil { _ = file.Close(); return fmt.Errorf("sincronizar snapshot temporal: %w", err) }
	if err = file.Close(); err != nil { return fmt.Errorf("cerrar snapshot temporal: %w", err) }
	if err = os.Rename(tmp, path); err != nil { return fmt.Errorf("confirmar snapshot: %w", err) }
	if dir, openErr := os.Open(rt.DataDir); openErr == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	ok = true
	return nil
}

// Restore valida versión, checksum y compatibilidad con el ensamblaje activo
// antes de modificar cualquier átomo, historial o auditoría.
// Nunca restaura capacidades ni RootCID desde el archivo: ambos proceden
// del ensamblaje declarado y se comparan contra el snapshot.
func (rt *Runtime) Restore() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.DataDir == "" { return fmt.Errorf("DataDir vacío") }
	if rt.Assembly == nil { return fmt.Errorf("LoadAssembly debe ejecutarse antes de Restore") }
	path := filepath.Join(rt.DataDir, "runtime_state.json")
	encoded, err := os.ReadFile(path)
	if err != nil { return err }

	snapshot, legacy, err := decodificarSnapshot(encoded)
	if err != nil { return err }
	if snapshot.SchemaVersion != VersionSnapshot && !legacy {
		return fmt.Errorf("versión de snapshot no soportada: %d", snapshot.SchemaVersion)
	}
	if !reflect.DeepEqual(snapshot.RootCIDs, rt.RootCIDs) {
		return fmt.Errorf("snapshot incompatible: RootCID de definición distinto al ensamblaje cargado")
	}
	if snapshot.AssemblyFingerprint != "" && snapshot.AssemblyFingerprint != huellaEnsamblaje(rt.RootCIDs) {
		return fmt.Errorf("snapshot incompatible: huella del ensamblaje no coincide")
	}
	if snapshot.Bindings != nil && !bindingsIgual(snapshot.Bindings, rt.Assembly.Bindings) {
		return fmt.Errorf("snapshot incompatible: bindings distintos")
	}
	// Una instantánea nunca puede elevar o sustituir la autoridad cargada.
	if snapshot.Caps != nil && !jsonIgual(snapshot.Caps, rt.Caps) {
		return fmt.Errorf("snapshot rechazado: capacidades distintas a las declaradas")
	}
	for _, record := range snapshot.Evaluations {
		if record.Schema == "" || record.RootOID == "" || record.RuleID == "" || record.State < 0 || record.State > 2 {
			return fmt.Errorf("snapshot contiene metadatos de evaluación inválidos")
		}
		sum := sha256.Sum256(record.Payload)
		if !json.Valid(record.Payload) || hex.EncodeToString(sum[:]) != record.Digest {
			return fmt.Errorf("snapshot contiene evaluación corrupta: %s/%s", record.RootOID, record.RuleID)
		}
		if _, ok := rt.Assembly.Orgs[record.RootOID]; !ok {
			return fmt.Errorf("snapshot contiene evaluación de organismo desconocido: %s", record.RootOID)
		}
	}

	state := ReactiveSnapshot{
		Atoms: snapshot.Atoms, Deps: snapshot.Deps, Dirty: snapshot.Dirty, History: snapshot.History,
	}
	// A partir de aquí no quedan operaciones de validación que puedan fallar.
	rt.State.ImportSnapshot(state)
	rt.Audit = append([]AuditEntry(nil), snapshot.Audit...)
	rt.Evaluations = copiarEvaluaciones(snapshot.Evaluations)
	rt.audit("restore", "", "snapshot validado y recuperado", fmt.Sprintf("version=%d legacy=%t", snapshot.SchemaVersion, legacy))
	return nil
}

func decodificarSnapshot(encoded []byte) (runtimeSnapshot, bool, error) {
	var shape map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &shape); err != nil {
		return runtimeSnapshot{}, false, fmt.Errorf("snapshot JSON inválido: %w", err)
	}
	if raw, ok := shape["payload"]; ok {
		var envelope snapshotEnvelope
		if err := json.Unmarshal(encoded, &envelope); err != nil { return runtimeSnapshot{}, false, err }
		if envelope.Version != VersionSnapshot { return runtimeSnapshot{}, false, fmt.Errorf("versión de snapshot no soportada: %d", envelope.Version) }
		if envelope.Algorithm != "sha256" { return runtimeSnapshot{}, false, fmt.Errorf("algoritmo de checksum no soportado: %s", envelope.Algorithm) }
		if !json.Valid(raw) { return runtimeSnapshot{}, false, fmt.Errorf("payload del snapshot inválido") }
		var snapshot runtimeSnapshot
		if err := json.Unmarshal(raw, &snapshot); err != nil { return runtimeSnapshot{}, false, fmt.Errorf("payload incompatible: %w", err) }
		canonical, err := json.Marshal(snapshot)
		if err != nil { return runtimeSnapshot{}, false, fmt.Errorf("canonicalizar payload: %w", err) }
		sum := sha256.Sum256(canonical)
		if hex.EncodeToString(sum[:]) != envelope.Checksum { return runtimeSnapshot{}, false, fmt.Errorf("checksum del snapshot no coincide") }
		if snapshot.SchemaVersion != envelope.Version { return runtimeSnapshot{}, false, fmt.Errorf("versión interna y externa del snapshot no coinciden") }
		return normalizarSnapshot(snapshot), false, nil
	}

	// Migración controlada del formato anterior sin envoltura. Se exige RootCID
	// para no aplicar datos históricos sobre una definición no identificada.
	var legacy struct {
		Atoms map[string]interface{} `json:"atoms"`
		RootCIDs map[string]string `json:"root_cids"`
		Bindings []Binding `json:"bindings"`
		Audit []AuditEntry `json:"audit"`
		History []InferenceRecord `json:"history"`
		Caps map[string]map[string]bool `json:"caps"`
	}
	if err := json.Unmarshal(encoded, &legacy); err != nil { return runtimeSnapshot{}, false, err }
	if legacy.RootCIDs == nil { return runtimeSnapshot{}, false, fmt.Errorf("snapshot legado sin RootCID: recuperación rechazada") }
	return normalizarSnapshot(runtimeSnapshot{
		SchemaVersion: 0, RootCIDs: legacy.RootCIDs, Bindings: legacy.Bindings,
		Atoms: legacy.Atoms, History: legacy.History, Audit: legacy.Audit, Caps: legacy.Caps,
	}), true, nil
}

func normalizarSnapshot(s runtimeSnapshot) runtimeSnapshot {
	if s.Atoms == nil { s.Atoms = map[string]interface{}{} }
	if s.Deps == nil { s.Deps = map[string][]string{} }
	if s.Dirty == nil { s.Dirty = map[string]bool{} }
	if s.History == nil { s.History = []InferenceRecord{} }
	if s.Audit == nil { s.Audit = []AuditEntry{} }
	if s.Evaluations == nil { s.Evaluations = []EvaluationRecord{} }
	return s
}

func huellaEnsamblaje(rootCIDs map[string]string) string {
	canonical, _ := json.Marshal(rootCIDs)
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:])
}

func copiarMapaString(source map[string]string) map[string]string {
	out := make(map[string]string, len(source))
	for key, value := range source { out[key] = value }
	return out
}

func copiarCapacidades(source map[string]map[string]bool) map[string]map[string]bool {
	out := make(map[string]map[string]bool, len(source))
	for oid, caps := range source {
		out[oid] = make(map[string]bool, len(caps))
		for capability, granted := range caps { out[oid][capability] = granted }
	}
	return out
}

func copiarEvaluaciones(source []EvaluationRecord) []EvaluationRecord {
	out := make([]EvaluationRecord, len(source))
	for i, record := range source {
		out[i] = record
		out[i].Payload = append(json.RawMessage(nil), record.Payload...)
	}
	return out
}

func bindingsIgual(a, b []Binding) bool {
	if len(a) != len(b) { return false }
	if len(a) == 0 { return true }
	return jsonIgual(a, b)
}

func jsonIgual(a, b interface{}) bool {
	left, err1 := json.Marshal(a)
	right, err2 := json.Marshal(b)
	return err1 == nil && err2 == nil && string(left) == string(right)
}
