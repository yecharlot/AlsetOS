// alset-desktop-bridge: local IPC + static Alset Shell for Tiny Core desktop.
// Does not replace the window manager; sits above FLWM and talks to AlsetOS.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type status struct {
	Service   string    `json:"service"`
	Time      time.Time `json:"time"`
	DataDir   string    `json:"data_dir"`
	Organism  string    `json:"organism"`
	RootCID   string    `json:"rootcid,omitempty"`
	AlsetOS   string    `json:"alsetos"` // running | missing | error
	Message   string    `json:"message,omitempty"`
}

type bridge struct {
	dataDir string
	mu      sync.Mutex
	orgName string
	rootCID string
	lastMsg string
}

func main() {
	addr := flag.String("addr", "127.0.0.1:7420", "bind address (localhost only recommended)")
	data := flag.String("data", "", "persistent data dir")
	shellDir := flag.String("shell", "", "path to desktop/shell static files")
	alsetos := flag.String("alsetos", "alsetos", "alsetos binary name/path for organism demo")
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

	sh := *shellDir
	if sh == "" {
		// walk common locations relative to cwd / executable
		cands := []string{
			"desktop/shell",
			filepath.Join("..", "desktop", "shell"),
		}
		if exe, err := os.Executable(); err == nil {
			cands = append([]string{filepath.Join(filepath.Dir(exe), "desktop", "shell")}, cands...)
		}
		for _, c := range cands {
			if st, err := os.Stat(filepath.Join(c, "index.html")); err == nil && !st.IsDir() {
				sh, _ = filepath.Abs(c)
				break
			}
		}
	}

	b := &bridge{dataDir: dataDir, orgName: "desktop-local"}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "service": "alset-desktop-bridge"})
	})
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, b.status(*alsetos))
	})
	mux.HandleFunc("/v1/organism/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			w.WriteHeader(405)
			return
		}
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
		// safety: only under dataDir
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

	if sh != "" {
		log.Printf("shell static: %s", sh)
		mux.Handle("/", http.FileServer(http.Dir(sh)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintf(w, "alset-desktop-bridge OK\nSet -shell path to desktop/shell\n")
		})
	}

	log.Printf("Alset Desktop Bridge on http://%s/", *addr)
	log.Printf("data dir: %s", dataDir)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func (b *bridge) status(alsetosBin string) status {
	b.mu.Lock()
	defer b.mu.Unlock()
	st := status{
		Service:  "alset-desktop-bridge",
		Time:     time.Now().UTC(),
		DataDir:  b.dataDir,
		Organism: b.orgName,
		RootCID:  b.rootCID,
		Message:  b.lastMsg,
	}
	if path, err := exec.LookPath(alsetosBin); err != nil {
		st.AlsetOS = "missing"
	} else {
		st.AlsetOS = "found:" + path
	}
	return st
}

func (b *bridge) runOrganism(alsetosBin string) (string, error) {
	// Prefer ejemplos/demo.alset if present
	manifest := "ejemplos/demo.alset"
	if _, err := os.Stat(manifest); err != nil {
		// write a minimal manifest into data dir
		manifest = filepath.Join(b.dataDir, "demo.alset")
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
		// best-effort parse RootCID line
		for _, line := range splitLines(string(out)) {
			if len(line) > 9 && line[:9] == "[ROOTCID]" {
				b.rootCID = trimSpace(line[9:])
			}
		}
	}
	b.mu.Unlock()
	return string(out), err
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
	return rel != ".." && (len(rel) < 2 || rel[:2] != "..")
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
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

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == ':') {
		s = s[1:]
	}
	return s
}
