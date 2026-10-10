// AlsetOS Native — núcleo ATS-001 en Go + entorno gráfico soberano (sin TinyCore/Node).
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yecharlot/AlsetOS/internal/ats001/kernel"
)

//go:embed ui
var uiFS embed.FS

func main() {
	addr := flag.String("addr", "127.0.0.1:7700", "listen address")
	data := flag.String("data", "", "data directory")
	flag.Parse()
	dataDir := *data
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		if home == "" {
			home = "."
		}
		dataDir = filepath.Join(home, ".alsetos-native")
	}
	k := kernel.New(dataDir)
	master := k.BootMaster()
	log.Printf("AlsetOS Native ATS-001")
	log.Printf("  Master OID=%s", master.OID())
	log.Printf("  RootCID=%s", master.RootCID())
	log.Printf("  UI http://%s/", *addr)

	sub, err := fs.Sub(uiFS, "ui")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, k.Status())
	})
	mux.HandleFunc("/api/tick", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		writeJSON(w, k.Tick())
	})
	mux.HandleFunc("/api/orges", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		var body struct {
			Name  string   `json:"name"`
			Kind  string   `json:"kind"`
			Goals []string `json:"goals"`
			Caps  []string `json:"caps"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if body.Name == "" {
			body.Name = body.Kind
		}
		if body.Kind == "" {
			body.Kind = "custom"
		}
		o, err := k.CreateORGES(k.MasterOID, body.Name, body.Kind, body.Goals, body.Caps)
		if err != nil {
			http.Error(w, err.Error(), 403)
			return
		}
		writeJSON(w, o.Public())
	})
	mux.HandleFunc("/api/pulse", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		var body struct {
			Source  string                 `json:"source"`
			Dest    string                 `json:"dest"`
			Payload map[string]interface{} `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		msg, err := k.Pulse(body.Source, body.Dest, body.Payload, nil)
		if err != nil {
			http.Error(w, err.Error(), 403)
			return
		}
		writeJSON(w, msg)
	})
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"ok": "true", "spec": "ATS-001", "product": "AlsetOS Native"})
	})

	fmt.Printf("\n  Abre en el navegador: http://%s/\n\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
