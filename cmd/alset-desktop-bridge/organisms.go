package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// orgKind: app | file | volume | network | user | process | ipfs | system
type sysOrganism struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"`
	Name      string         `json:"name"`
	RootCID   string         `json:"rootcid"`
	State     string         `json:"state"`
	Meta      map[string]any `json:"meta,omitempty"`
	Created   time.Time      `json:"created"`
	Updated   time.Time      `json:"updated"`
}

type orgRegistry struct {
	mu   sync.Mutex
	path string
	All  map[string]*sysOrganism `json:"organisms"`
}

func newOrgRegistry(dataDir string) *orgRegistry {
	r := &orgRegistry{
		path: filepath.Join(dataDir, "organisms.json"),
		All:  map[string]*sysOrganism{},
	}
	r.load()
	return r
}

func (r *orgRegistry) load() {
	raw, err := os.ReadFile(r.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(raw, r)
	if r.All == nil {
		r.All = map[string]*sysOrganism{}
	}
}

func (r *orgRegistry) save() {
	r.mu.Lock()
	defer r.mu.Unlock()
	raw, _ := json.MarshalIndent(struct {
		Organisms map[string]*sysOrganism `json:"organisms"`
	}{r.All}, "", "  ")
	_ = os.WriteFile(r.path, raw, 0o644)
}

func (r *orgRegistry) upsert(o *sysOrganism) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o.Updated = time.Now().UTC()
	if o.Created.IsZero() {
		o.Created = o.Updated
	}
	if o.RootCID == "" {
		o.RootCID = "cid:org:" + o.Kind + ":" + o.ID
	}
	r.All[o.ID] = o
}

func (b *bridge) refreshSystemOrganisms() {
	// volumes
	for _, root := range []string{"/media", "/mnt", b.dataDir} {
		if st, err := os.Stat(root); err == nil && st.IsDir() {
			id := "vol:" + strings.ReplaceAll(root, "/", "_")
			b.orgs.upsert(&sysOrganism{
				ID: id, Kind: "volume", Name: root, State: "mounted",
				Meta: map[string]any{"path": root, "local": true},
			})
			entries, _ := os.ReadDir(root)
			for _, e := range entries {
				if e.IsDir() {
					p := filepath.Join(root, e.Name())
					b.orgs.upsert(&sysOrganism{
						ID: "vol:" + e.Name(), Kind: "volume", Name: e.Name(), State: "mounted",
						Meta: map[string]any{"path": p, "removable_guess": root == "/media" || root == "/mnt"},
					})
				}
			}
		}
	}
	// network interfaces
	ifs, _ := net.Interfaces()
	for _, iface := range ifs {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		var as []string
		for _, a := range addrs {
			as = append(as, a.String())
		}
		up := iface.Flags&net.FlagUp != 0
		st := "down"
		if up {
			st = "up"
		}
		kind := "network"
		name := iface.Name
		meta := map[string]any{"addrs": as, "mtu": iface.MTU, "flags": iface.Flags.String()}
		if strings.HasPrefix(name, "wl") || strings.Contains(name, "wlan") {
			meta["medium"] = "wifi"
		} else if strings.HasPrefix(name, "e") || strings.Contains(name, "eth") {
			meta["medium"] = "ethernet"
		} else {
			meta["medium"] = "other"
		}
		b.orgs.upsert(&sysOrganism{
			ID: "net:" + name, Kind: kind, Name: name, State: st, Meta: meta,
		})
	}
	b.orgs.save()
}

func (b *bridge) handleOrganismsList(w http.ResponseWriter, r *http.Request) {
	if _, err := b.auth.require(r, roleGuest); err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	b.refreshSystemOrganisms()
	kind := r.URL.Query().Get("kind")
	b.orgs.mu.Lock()
	list := make([]*sysOrganism, 0, len(b.orgs.All))
	for _, o := range b.orgs.All {
		if kind != "" && o.Kind != kind {
			continue
		}
		list = append(list, o)
	}
	b.orgs.mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "organisms": list, "model": "everything-is-organism"})
}

func (b *bridge) handleOrganismsGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	b.orgs.mu.Lock()
	o := b.orgs.All[id]
	b.orgs.mu.Unlock()
	if o == nil {
		http.Error(w, "not found", 404)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "organism": o})
}

func (b *bridge) handleVolumes(w http.ResponseWriter, r *http.Request) {
	b.refreshSystemOrganisms()
	b.orgs.mu.Lock()
	var vols []*sysOrganism
	for _, o := range b.orgs.All {
		if o.Kind == "volume" {
			vols = append(vols, o)
		}
	}
	b.orgs.mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "volumes": vols})
}

func (b *bridge) handleNetwork(w http.ResponseWriter, r *http.Request) {
	b.refreshSystemOrganisms()
	b.orgs.mu.Lock()
	var nets []*sysOrganism
	for _, o := range b.orgs.All {
		if o.Kind == "network" {
			nets = append(nets, o)
		}
	}
	b.orgs.mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "interfaces": nets})
}

// IPFS-like content store (local content-addressed blobs under dataDir/ipfs)
func (b *bridge) handleIPFSAdd(w http.ResponseWriter, r *http.Request) {
	if _, err := b.auth.require(r, roleUser); err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	var body struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 4<<20)).Decode(&body)
	sum := sha256Hex(body.Content)
	cid := "cid:ipfs:" + sum[:16]
	dir := filepath.Join(b.dataDir, "ipfs")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, sum+".bin"), []byte(body.Content), 0o644)
	meta := map[string]any{"name": body.Name, "cid": cid, "sha256": sum, "size": len(body.Content)}
	raw, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, sum+".json"), raw, 0o644)
	b.orgs.upsert(&sysOrganism{
		ID: "ipfs:" + sum[:12], Kind: "ipfs", Name: body.Name, RootCID: cid, State: "stored",
		Meta: meta,
	})
	b.orgs.save()
	b.audit("master", "ipfs.add", cid)
	writeJSON(w, map[string]any{"ok": true, "cid": cid, "sha256": sum})
}

func (b *bridge) handleIPFSList(w http.ResponseWriter, r *http.Request) {
	dir := filepath.Join(b.dataDir, "ipfs")
	_ = os.MkdirAll(dir, 0o755)
	entries, _ := os.ReadDir(dir)
	var items []map[string]any
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			var m map[string]any
			if json.Unmarshal(raw, &m) == nil {
				items = append(items, m)
			}
		}
	}
	writeJSON(w, map[string]any{"ok": true, "items": items})
}

func (b *bridge) handleIPFSGet(w http.ResponseWriter, r *http.Request) {
	cid := r.URL.Query().Get("cid")
	sha := strings.TrimPrefix(cid, "cid:ipfs:")
	// find by prefix
	dir := filepath.Join(b.dataDir, "ipfs")
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".bin") && strings.HasPrefix(e.Name(), sha) {
			raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			writeJSON(w, map[string]any{"ok": true, "content": string(raw)})
			return
		}
	}
	// full sha file
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".bin") {
			name := strings.TrimSuffix(e.Name(), ".bin")
			if strings.HasPrefix(name, sha) || strings.Contains(cid, name[:min(16, len(name))]) {
				raw, _ := os.ReadFile(filepath.Join(dir, e.Name()))
				writeJSON(w, map[string]any{"ok": true, "content": string(raw)})
				return
			}
		}
	}
	http.Error(w, "not found", 404)
}

func sha256Hex(s string) string {
	sum := sha256Sum([]byte(s))
	return sum
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
