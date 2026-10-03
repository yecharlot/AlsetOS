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
	mux.HandleFunc("/v1/fs/list", func(w http.ResponseWriter, r *http.Request) {
		dir := r.URL.Query().Get("path")
		if dir == "" {
			dir = dataDir
		}
		abs, err := filepath.Abs(dir)
		if err != nil || !under(dataDir, abs) {
			http.Error(w, "path not allowed", 403)
			return
		}
		entries, err := os.ReadDir(abs)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		writeJSON(w, map[string]any{"path": abs, "entries": names})
	})

	// —— Cognition (system services for OS + apps) ——
	mux.HandleFunc("/v1/mind/tick", b.handleMind)
	mux.HandleFunc("/v1/zyrion", b.handleZyrion)
	mux.HandleFunc("/v1/syllogism/assert", b.handleAssert)
	mux.HandleFunc("/v1/syllogism/infer", b.handleInfer)
	mux.HandleFunc("/v1/syllogism/ask", b.handleAsk)
	mux.HandleFunc("/v1/neural", b.handleNeural)

	if web != "" {
		log.Printf("tools web: %s", web)
		mux.Handle("/tools/", http.StripPrefix("/tools/", http.FileServer(http.Dir(web))))
		mux.Handle("/studio/", http.StripPrefix("/studio/", http.FileServer(http.Dir(web))))
		mux.Handle("/alset-editor/", http.StripPrefix("/alset-editor/", http.FileServer(http.Dir(filepath.Join(web, "alset-editor")))))
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
