// alset-desktop-bridge: windowed desktop IPC + Studio/Editor + Mind/Zyrion/Neural/Syllogism.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yecharlot/AlsetOS/mind"
	"github.com/yecharlot/AlsetOS/organismo"
	"github.com/yecharlot/AlsetOS/zyrion"
)

type status struct {
	Service  string    `json:"service"`
	Time     time.Time `json:"time"`
	DataDir  string    `json:"data_dir"`
	Organism string    `json:"organism"`
	RootCID  string    `json:"rootcid,omitempty"`
	AlsetOS  string    `json:"alsetos"`
	Tools    string    `json:"tools,omitempty"`
	Message  string    `json:"message,omitempty"`
	Cognition string   `json:"cognition"`
}

type fact struct {
	S string  `json:"s"`
	R string  `json:"r"`
	O string  `json:"o"`
	C float64 `json:"c"`
}

type bridge struct {
	dataDir string
	webDir  string
	mu      sync.Mutex
	orgName string
	rootCID string
	lastMsg string
	facts   []fact
	neural  map[string]float64
	mente   mind.Mente
}

func main() {
	addr := flag.String("addr", "127.0.0.1:7420", "bind address")
	data := flag.String("data", "", "persistent data dir")
	shellDir := flag.String("shell", "", "desktop/shell static")
	webDir := flag.String("web", "", "Alset-LISPAI-Runtime/web")
	alsetos := flag.String("alsetos", "alsetos", "alsetos binary")
	flag.Parse()

	dataDir := *data
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		if home == "" {
			home = "/tmp"
		}
		dataDir = filepath.Join(home, ".alset-desktop")
	}
	_ = os.MkdirAll(dataDir, 0o755)

	sh := resolveDir(*shellDir, []string{"desktop/shell", filepath.Join("..", "desktop", "shell")})
	web := *webDir
	if web == "" {
		web = resolveDir("", []string{
			filepath.Join("..", "Alset-LISPAI-Runtime", "web"),
			filepath.Join("..", "..", "Alset-LISPAI-Runtime", "web"),
		})
	}

	b := &bridge{
		dataDir: dataDir,
		webDir:  web,
		orgName: "desktop-local",
		neural:  map[string]float64{"atencion.ui": 0.5, "preferencia.studio": 0.6},
		facts: []fact{
			{S: "organismo", R: "tiene_capacidad", O: "backup", C: 1},
			{S: "alsetos", R: "es", O: "sistema", C: 1},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "service": "alset-desktop-bridge", "tools": web != "", "cognition": true})
	})
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, b.status(*alsetos))
	})
	mux.HandleFunc("/v1/organism/run", func(w http.ResponseWriter, r *http.Request) {
		out, err := b.runOrganism(*alsetos)
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "error": err.Error(), "output": out})
			return
		}
		writeJSON(w, map[string]any{"ok": true, "output": out, "organism": b.orgName, "rootcid": b.rootCID})
	})
	mux.HandleFunc("/v1/fs/list", b.handleFSList)
	mux.HandleFunc("/v1/fs/read", b.handleFSRead)
	mux.HandleFunc("/v1/fs/write", b.handleFSWrite)
	mux.HandleFunc("/v1/fs/mkdir", b.handleFSMkdir)
	mux.HandleFunc("/v1/apps/list", b.handleAppsList)
	mux.HandleFunc("/v1/apps/deploy", b.handleAppsDeploy)
	mux.HandleFunc("/v1/apps/get", b.handleAppsGet)
	// Installed apps as static files under /apps/<name>/
	appsDir := filepath.Join(dataDir, "apps")
	_ = os.MkdirAll(appsDir, 0o755)
	mux.Handle("/apps/", http.StripPrefix("/apps/", http.FileServer(http.Dir(appsDir))))

	// —— Cognition (system services for OS + apps) ——
	mux.HandleFunc("/v1/mind/tick", b.handleMind)
	mux.HandleFunc("/v1/zyrion", b.handleZyrion)
	mux.HandleFunc("/v1/syllogism/assert", b.handleAssert)
	mux.HandleFunc("/v1/syllogism/infer", b.handleInfer)
	mux.HandleFunc("/v1/syllogism/ask", b.handleAsk)
	mux.HandleFunc("/v1/neural", b.handleNeural)

	if web != "" {
		log.Printf("tools web: %s", web)
		// Absolute paths expected by Studio/Editor (same as standalone :5177)
		mount := func(urlPath, dir string) {
			if st, err := os.Stat(dir); err == nil && st.IsDir() {
				mux.Handle(urlPath, http.StripPrefix(strings.TrimSuffix(urlPath, "/"), http.FileServer(http.Dir(dir))))
				// also with trailing patterns via FileServer parent
			}
		}
		// Serve full web tree at /tools/ (Studio index)
		mux.Handle("/tools/", http.StripPrefix("/tools/", http.FileServer(http.Dir(web))))
		// Root-absolute asset mounts so /css /studio /runtime resolve inside desktop
		for _, sub := range []string{"css", "studio", "runtime", "alset", "lispai", "alset-editor"} {
			dir := filepath.Join(web, sub)
			if st, err := os.Stat(dir); err == nil && st.IsDir() {
				pfx := "/" + sub
				mux.Handle(pfx+"/", http.StripPrefix(pfx, http.FileServer(http.Dir(dir))))
				log.Printf("mount %s/ → %s", pfx, dir)
			}
		}
		// Studio index also at /studio-app/ for clarity
		mux.HandleFunc("/studio-app/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(web, "index.html"))
		})
		_ = mount
	}
	if sh != "" {
		log.Printf("shell static: %s", sh)
		mux.Handle("/", http.FileServer(http.Dir(sh)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "alset-desktop-bridge OK\n")
		})
	}

	log.Printf("Alset Desktop http://%s/  (ventana + Mind/Zyrion)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func (b *bridge) handleMind(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text string `json:"text"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body)
	text := strings.TrimSpace(body.Text)
	if text == "" {
		text = "pulso"
	}

	// Ternary disposition from keyword heuristics + organism capability
	estado := zyrion.Incierto
	low := strings.ToLower(text)
	switch {
	case strings.Contains(low, "si") || strings.Contains(low, "ejecut") || strings.Contains(low, "listo"):
		estado = zyrion.Si
	case strings.Contains(low, "no") || strings.Contains(low, "detener") || strings.Contains(low, "error"):
		estado = zyrion.No
	}

	ent := &organismo.Organismo{
		Nombre:    b.orgName,
		Capacidad: map[string]bool{"gene.ejecutar": true},
		Memoria:   map[string]string{"evaluacion": "si", "ultimo_texto": text},
	}
	decision := b.mente.DecidirConEstrategia(estado, ent, mind.Evaluar)

	writeJSON(w, map[string]any{
		"ok":       true,
		"voice":    fmt.Sprintf("observé «%s» → zyrion=%s → mind=%s", truncate(text, 48), estado.Texto(), decision),
		"text":     text,
		"estado":   estado.Texto(),
		"decision": decision,
		"organism": b.orgName,
		"at":       time.Now().UTC(),
	})
}

func (b *bridge) handleZyrion(w http.ResponseWriter, r *http.Request) {
	var body struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body)
	// Map continuous signals to ternary via thresholds
	score := (body.A + body.B) / 2
	var v zyrion.Valor
	switch {
	case score >= 0.66:
		v = zyrion.Si
	case score <= 0.33:
		v = zyrion.No
	default:
		v = zyrion.Incierto
	}
	writeJSON(w, map[string]any{
		"ok":     true,
		"a":      body.A,
		"b":      body.B,
		"score":  score,
		"valor":  int(v),
		"estado": v.Texto(),
		"engine": "alsetos/zyrion",
	})
}

func (b *bridge) handleAssert(w http.ResponseWriter, r *http.Request) {
	var f fact
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&f)
	if f.S == "" || f.R == "" {
		http.Error(w, "s and r required", 400)
		return
	}
	if f.C == 0 {
		f.C = 1
	}
	b.mu.Lock()
	b.facts = append(b.facts, f)
	n := len(b.facts)
	b.mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "asserted": f, "n": n})
}

func (b *bridge) handleInfer(w http.ResponseWriter, r *http.Request) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Simple chain: if A-rel-B and B-rel-C → A-rel-C (weak)
	derived := []fact{}
	for i := 0; i < len(b.facts); i++ {
		for j := 0; j < len(b.facts); j++ {
			if i == j {
				continue
			}
			a, c := b.facts[i], b.facts[j]
			if a.O == c.S && a.R == c.R {
				derived = append(derived, fact{S: a.S, R: a.R, O: c.O, C: math.Min(a.C, c.C) * 0.8})
			}
		}
	}
	writeJSON(w, map[string]any{"ok": true, "facts": b.facts, "derived": derived})
}

func (b *bridge) handleAsk(w http.ResponseWriter, r *http.Request) {
	var q fact
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&q)
	b.mu.Lock()
	defer b.mu.Unlock()
	matches := []fact{}
	for _, f := range b.facts {
		if (q.S == "" || f.S == q.S) && (q.R == "" || f.R == q.R) && (q.O == "" || f.O == q.O) {
			matches = append(matches, f)
		}
	}
	writeJSON(w, map[string]any{"ok": true, "matches": matches, "count": len(matches)})
}

func (b *bridge) handleNeural(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		key := r.URL.Query().Get("key")
		b.mu.Lock()
		defer b.mu.Unlock()
		if key != "" {
			writeJSON(w, map[string]any{"ok": true, "key": key, "value": b.neural[key]})
			return
		}
		writeJSON(w, map[string]any{"ok": true, "weights": b.neural})
		return
	}
	var body struct {
		Key   string  `json:"key"`
		Value float64 `json:"value"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body)
	if body.Key == "" {
		http.Error(w, "key required", 400)
		return
	}
	v := body.Value
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	b.mu.Lock()
	b.neural[body.Key] = v
	b.mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "key": body.Key, "value": v})
}

func (b *bridge) status(alsetosBin string) status {
	b.mu.Lock()
	defer b.mu.Unlock()
	st := status{
		Service:   "alset-desktop-bridge",
		Time:      time.Now().UTC(),
		DataDir:   b.dataDir,
		Organism:  b.orgName,
		RootCID:   b.rootCID,
		Message:   b.lastMsg,
		Tools:     b.webDir,
		Cognition: "mind+zyrion+neural+syllogism",
	}
	if path, err := exec.LookPath(alsetosBin); err == nil {
		st.AlsetOS = "found:" + path
	} else if _, err2 := os.Stat(alsetosBin); err2 == nil {
		st.AlsetOS = "found:" + alsetosBin
	} else {
		st.AlsetOS = "missing"
	}
	return st
}

func (b *bridge) runOrganism(alsetosBin string) (string, error) {
	manifest := filepath.Join(b.dataDir, "demo.alset")
	if _, err := os.Stat(manifest); err != nil {
		_ = os.WriteFile(manifest, []byte(`{
  "nombre": "desktop-local",
  "zyrion": { "estado": "si" },
  "capacidades": ["gene.ejecutar"],
  "genes": ["gene-saludo"]
}`), 0o644)
	}
	cmd := exec.Command(alsetosBin, manifest)
	cmd.Dir = findModuleRoot()
	out, err := cmd.CombinedOutput()
	b.mu.Lock()
	b.lastMsg = string(out)
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "[ROOTCID]") {
				b.rootCID = strings.TrimSpace(strings.TrimPrefix(line, "[ROOTCID]"))
				b.rootCID = strings.TrimLeft(b.rootCID, ": ")
			}
		}
	}
	b.mu.Unlock()
	return string(out), err
}


func (b *bridge) safePath(rel string) (string, error) {
	if rel == "" || rel == "." {
		return b.dataDir, nil
	}
	// allow absolute only if under dataDir
	var candidate string
	if filepath.IsAbs(rel) {
		candidate = rel
	} else {
		candidate = filepath.Join(b.dataDir, rel)
	}
	abs, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	if !under(b.dataDir, abs) {
		return "", fmt.Errorf("path not allowed")
	}
	return abs, nil
}

func (b *bridge) handleFSList(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	abs, err := b.safePath(rel)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	type ent struct {
		Name  string `json:"name"`
		Dir   bool   `json:"dir"`
		Size  int64  `json:"size,omitempty"`
	}
	var list []ent
	for _, e := range entries {
		item := ent{Name: e.Name(), Dir: e.IsDir()}
		if fi, err := e.Info(); err == nil && !e.IsDir() {
			item.Size = fi.Size()
		}
		list = append(list, item)
	}
	parent := ""
	if abs != b.dataDir {
		parent = filepath.Dir(abs)
		if !under(b.dataDir, parent) {
			parent = b.dataDir
		}
		// relative parent for UI
		if relp, err := filepath.Rel(b.dataDir, parent); err == nil {
			parent = relp
		}
	}
	curRel, _ := filepath.Rel(b.dataDir, abs)
	writeJSON(w, map[string]any{"path": abs, "rel": curRel, "parent": parent, "root": b.dataDir, "entries": list})
}

func (b *bridge) handleFSRead(w http.ResponseWriter, r *http.Request) {
	abs, err := b.safePath(r.URL.Query().Get("path"))
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	if len(data) > 512*1024 {
		http.Error(w, "file too large", 413)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "path": abs, "content": string(data)})
}

func (b *bridge) handleFSWrite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var body struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 2<<20)).Decode(&body)
	abs, err := b.safePath(body.Path)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	_ = os.MkdirAll(filepath.Dir(abs), 0o755)
	if err := os.WriteFile(abs, []byte(body.Content), 0o644); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "path": abs})
}

func (b *bridge) handleFSMkdir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body)
	abs, err := b.safePath(body.Path)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "path": abs})
}

func (b *bridge) handleAppsList(w http.ResponseWriter, r *http.Request) {
	dir := filepath.Join(b.dataDir, "apps")
	_ = os.MkdirAll(dir, 0o755)
	entries, _ := os.ReadDir(dir)
	var apps []map[string]any
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		meta := map[string]any{"name": e.Name(), "url": "/apps/" + e.Name() + "/"}
		if raw, err := os.ReadFile(filepath.Join(dir, e.Name(), "app.json")); err == nil {
			var m map[string]any
			if json.Unmarshal(raw, &m) == nil {
				for k, v := range m {
					meta[k] = v
				}
			}
		}
		apps = append(apps, meta)
	}
	writeJSON(w, map[string]any{"ok": true, "apps": apps})
}

func (b *bridge) handleAppsDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var body struct {
		Name  string `json:"name"`
		Title string `json:"title"`
		HTML  string `json:"html"`
		Kind  string `json:"kind"` // html | alset-js
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 4<<20)).Decode(&body)
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "app"
	}
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)
	dir := filepath.Join(b.dataDir, "apps", name)
	_ = os.MkdirAll(dir, 0o755)
	html := body.HTML
	if html == "" {
		// default calculator demo
		html = `<!DOCTYPE html><html><head><meta charset="utf-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>Calculadora</title>
<style>
body{margin:0;font-family:system-ui;background:#0c1018;color:#f2f4f8;display:flex;justify-content:center;padding:24px}
.calc{width:260px;background:#141824;border-radius:16px;padding:16px;border:1px solid rgba(255,255,255,.1)}
.display{background:#080a0e;padding:16px;border-radius:10px;text-align:right;font-size:28px;margin-bottom:12px;min-height:40px}
.grid{display:grid;grid-template-columns:repeat(4,1fr);gap:8px}
button{padding:14px;border:0;border-radius:10px;background:#1c2436;color:#fff;font-size:16px;cursor:pointer}
button.op{background:#d4a017;color:#111}
button.eq{background:#3ddc97;color:#111}
</style></head><body>
<div class="calc"><div class="display" id="d">0</div><div class="grid" id="g"></div></div>
<script>
let cur='0',op=null,acc=null;
const d=document.getElementById('d');
const keys=['7','8','9','/','4','5','6','*','1','2','3','-','0','.','=','+'];
keys.forEach(k=>{
  const b=document.createElement('button');
  b.textContent=k;
  if('/+-*'.includes(k))b.className='op';
  if(k==='=')b.className='eq';
  b.onclick=()=>{
    if('0123456789.'.includes(k)){cur=cur==='0'&&k!=='.'?k:cur+k;d.textContent=cur;return;}
    if(k==='='){if(op&&acc!=null){const n=Number(cur);let r=n;if(op==='+')r=acc+n;if(op==='-')r=acc-n;if(op==='*')r=acc*n;if(op==='/')r=acc/n;cur=String(r);acc=null;op=null;d.textContent=cur;}return;}
    acc=Number(cur);op=k;cur='0';
  };
  document.getElementById('g').appendChild(b);
});
</script></body></html>`
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(html), 0o644); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	meta, _ := json.MarshalIndent(map[string]any{
		"name": name, "title": body.Title, "kind": body.Kind, "deployed": time.Now().UTC(),
	}, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "app.json"), meta, 0o644)
	writeJSON(w, map[string]any{"ok": true, "name": name, "url": "/apps/" + name + "/", "title": body.Title})
}

func (b *bridge) handleAppsGet(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	dir := filepath.Join(b.dataDir, "apps", name)
	raw, err := os.ReadFile(filepath.Join(dir, "app.json"))
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}


func resolveDir(explicit string, cands []string) string {
	if explicit != "" {
		if st, err := os.Stat(explicit); err == nil && st.IsDir() {
			a, _ := filepath.Abs(explicit)
			return a
		}
	}
	if exe, err := os.Executable(); err == nil {
		cands = append([]string{filepath.Join(filepath.Dir(exe), "desktop", "shell")}, cands...)
	}
	for _, c := range cands {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			a, _ := filepath.Abs(c)
			return a
		}
	}
	return ""
}

func findModuleRoot() string {
	wd, _ := os.Getwd()
	for d := wd; d != "/" && d != ""; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
	}
	return wd
}

func under(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
