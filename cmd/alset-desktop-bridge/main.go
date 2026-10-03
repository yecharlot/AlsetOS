// alset-desktop-bridge: IPC + Alset Shell + optional Studio/Editor static tools.
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
	"strings"
	"sync"
	"time"
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
}

type bridge struct {
	dataDir string
	webDir  string
	mu      sync.Mutex
	orgName string
	rootCID string
	lastMsg string
}

func main() {
	addr := flag.String("addr", "127.0.0.1:7420", "bind address")
	data := flag.String("data", "", "persistent data dir")
	shellDir := flag.String("shell", "", "desktop/shell static")
	webDir := flag.String("web", "", "Alset-LISPAI-Runtime/web (Studio + Editor)")
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

	b := &bridge{dataDir: dataDir, webDir: web, orgName: "desktop-local"}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "service": "alset-desktop-bridge", "tools": web != ""})
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

	// Terminal helper: eval simple recordar via API
	mux.HandleFunc("/v1/term/echo", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeJSON(w, map[string]any{"ok": true, "echo": r.FormValue("q")})
	})

	if web != "" {
		log.Printf("tools web: %s", web)
		mux.Handle("/tools/", http.StripPrefix("/tools/", http.FileServer(http.Dir(web))))
		// aliases
		mux.Handle("/studio/", http.StripPrefix("/studio/", http.FileServer(http.Dir(filepath.Join(web)))))
		mux.Handle("/alset-editor/", http.StripPrefix("/alset-editor/", http.FileServer(http.Dir(filepath.Join(web, "alset-editor")))))
	}

	if sh != "" {
		log.Printf("shell static: %s", sh)
		mux.Handle("/", http.FileServer(http.Dir(sh)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "alset-desktop-bridge OK\nUse -shell desktop/shell\n")
		})
	}

	log.Printf("Alset Desktop Bridge http://%s/", *addr)
	log.Printf("data: %s", dataDir)
	if web != "" {
		log.Printf("Studio  http://%s/tools/  or /studio/", *addr)
		log.Printf("Editor  http://%s/tools/alset-editor/", *addr)
	}
	log.Fatal(http.ListenAndServe(*addr, mux))
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
		Tools:    b.webDir,
	}
	if path, err := exec.LookPath(alsetosBin); err != nil {
		if _, err2 := os.Stat(alsetosBin); err2 == nil {
			st.AlsetOS = "found:" + alsetosBin
		} else {
			st.AlsetOS = "missing"
		}
	} else {
		st.AlsetOS = "found:" + path
	}
	return st
}

func (b *bridge) runOrganism(alsetosBin string) (string, error) {
	manifest := "ejemplos/demo.alset"
	if _, err := os.Stat(manifest); err != nil {
		manifest = filepath.Join(b.dataDir, "demo.alset")
		_ = os.WriteFile(manifest, []byte(`{
  "nombre": "desktop-local",
  "zyrion": { "estado": "si" },
  "capacidades": ["gene.ejecutar"],
  "genes": ["gene-saludo"]
}`), 0o644)
	}
	bin := alsetosBin
	cmd := exec.Command(bin, manifest)
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

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
