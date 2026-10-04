package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

// Minimal MCP-compatible tools endpoint (JSON) so external agents can discover AlsetOS.
func (b *bridge) handleMCPTools(w http.ResponseWriter, r *http.Request) {
	tools := []map[string]any{
		{"name": "alset.status", "description": "Estado del bridge y cognición", "inputSchema": map[string]any{"type": "object"}},
		{"name": "alset.organisms.list", "description": "Lista organismos del sistema", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"kind": map[string]any{"type": "string"}}}},
		{"name": "alset.mind.tick", "description": "Latido Mind", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"text": map[string]any{"type": "string"}}}},
		{"name": "alset.apps.list", "description": "Apps instaladas", "inputSchema": map[string]any{"type": "object"}},
		{"name": "alset.ipfs.list", "description": "Contenido IPFS local", "inputSchema": map[string]any{"type": "object"}},
		{"name": "alset.fs.list", "description": "Listar archivos data dir", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"path": map[string]any{"type": "string"}}}},
	}
	writeJSON(w, map[string]any{"ok": true, "tools": tools, "protocol": "alset-mcp-lite"})
}

func (b *bridge) handleMCPCall(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body)
	b.audit("mcp", "mcp.call", body.Name)
	switch body.Name {
	case "alset.status":
		writeJSON(w, b.status("alsetos"))
	case "alset.organisms.list":
		b.refreshSystemOrganisms()
		b.orgs.mu.Lock()
		list := make([]*sysOrganism, 0, len(b.orgs.All))
		for _, o := range b.orgs.All {
			list = append(list, o)
		}
		b.orgs.mu.Unlock()
		writeJSON(w, map[string]any{"organisms": list})
	case "alset.apps.list":
		b.handleAppsList(w, r)
	case "alset.ipfs.list":
		b.handleIPFSList(w, r)
	case "alset.fs.list":
		// reuse by setting query — simplified
		path, _ := body.Arguments["path"].(string)
		abs, err := b.safePath(path)
		if err != nil {
			http.Error(w, err.Error(), 403)
			return
		}
		entries, _ := os.ReadDir(abs)
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		writeJSON(w, map[string]any{"path": abs, "entries": names})
	case "alset.mind.tick":
		text, _ := body.Arguments["text"].(string)
		writeJSON(w, map[string]any{"ok": true, "echo": text, "note": "use POST /v1/mind/tick"})
	default:
		http.Error(w, "unknown tool", 404)
	}
}

func jsonReader(v any) io.Reader {
	raw, _ := json.Marshal(v)
	return &bytesReader{b: raw}
}

type bytesReader struct {
	b []byte
	i int
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}

