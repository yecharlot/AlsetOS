package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type auditEntry struct {
	At        time.Time `json:"at"`
	AccountID string    `json:"account_id"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
}

type auditLog struct {
	mu   sync.Mutex
	path string
	Max  int
}

func newAuditLog(dataDir string) *auditLog {
	return &auditLog{path: filepath.Join(dataDir, "audit.jsonl"), Max: 5000}
}

func (b *bridge) audit(accountID, action, detail string) {
	if b.auditLog == nil {
		return
	}
	e := auditEntry{At: time.Now().UTC(), AccountID: accountID, Action: action, Detail: detail}
	raw, _ := json.Marshal(e)
	b.auditLog.mu.Lock()
	defer b.auditLog.mu.Unlock()
	f, err := os.OpenFile(b.auditLog.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(raw, '\n'))
	_ = f.Close()
	// Mind learns: store last action in neural/meta
	b.mu.Lock()
	b.neural["last.action"] = float64(len(action)%100) / 100
	b.mu.Unlock()
}

func (b *bridge) handleAuditList(w http.ResponseWriter, r *http.Request) {
	if _, err := b.auth.require(r, roleUser); err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	raw, err := os.ReadFile(b.auditLog.path)
	if err != nil {
		writeJSON(w, map[string]any{"ok": true, "entries": []any{}})
		return
	}
	lines := splitLines(string(raw))
	if len(lines) > 200 {
		lines = lines[len(lines)-200:]
	}
	var entries []json.RawMessage
	for _, ln := range lines {
		if ln == "" {
			continue
		}
		entries = append(entries, json.RawMessage(ln))
	}
	writeJSON(w, map[string]any{"ok": true, "entries": entries})
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func sha256Sum(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
