// AlsetOS Native — núcleo ATS-001 en Go + entorno gráfico soberano (sin TinyCore/Node).
package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// Export ORGES as static web + PWA (zip)
	mux.HandleFunc("/api/orges/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		var body struct {
			Name   string   `json:"name"`
			Kind   string   `json:"kind"`
			Title  string   `json:"title"`
			Goals  []string `json:"goals"`
			Caps   []string `json:"caps"`
			Color  string   `json:"color"`
			OID    string   `json:"oid"`
			Target string   `json:"target"` // web | pwa | android-stub
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if body.Title == "" {
			body.Title = body.Name
		}
		if body.Color == "" {
			body.Color = "#0d9488"
		}
		if body.Target == "" {
			body.Target = "pwa"
		}
		// Optionally register organism if not yet
		if body.OID == "" && body.Kind != "" {
			o, err := k.CreateORGES(k.MasterOID, body.Name, body.Kind, body.Goals, body.Caps)
			if err == nil {
				body.OID = o.OID()
			}
		}
		zipBytes, err := buildORGESPWA(body.Title, body.Kind, body.OID, body.Goals, body.Color)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// also store under data dir
		outDir := filepath.Join(dataDir, "exports")
		_ = os.MkdirAll(outDir, 0o755)
		fname := fmt.Sprintf("orges-%s-%d.zip", sanitize(body.Kind), time.Now().Unix())
		_ = os.WriteFile(filepath.Join(outDir, fname), zipBytes, 0o644)

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fname+"\"")
		_, _ = w.Write(zipBytes)
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

func sanitize(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}

func buildORGESPWA(title, kind, oid string, goals []string, color string) ([]byte, error) {
	goalsJSON, _ := json.Marshal(goals)
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<meta name="theme-color" content="%s"/>
<link rel="manifest" href="manifest.json"/>
<title>%s</title>
<style>
body{margin:0;font-family:system-ui,sans-serif;background:#0b1220;color:#eef4ff}
header{padding:16px 20px;background:linear-gradient(90deg,%s,#1e3a5f);font-weight:700}
main{padding:16px;display:grid;gap:12px;grid-template-columns:repeat(auto-fill,minmax(140px,1fr))}
.card{background:#141e30;border:1px solid rgba(120,160,220,.2);border-radius:12px;padding:14px}
.card b{display:block;font-size:11px;color:#8fa3c4;margin-bottom:6px}
.oid{font-family:monospace;font-size:11px;color:#5b9dff;word-break:break-all;padding:12px 20px}
</style>
</head>
<body>
<header>%s · ORGES</header>
<div class="oid">OID %s · kind %s</div>
<main id="kpis"></main>
<script>
const goals = %s;
document.getElementById('kpis').innerHTML = goals.map(g =>
  '<div class="card"><b>'+g+'</b><span>—</span></div>').join('') || '<div class="card"><b>estado</b>activo</div>';
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('sw.js').catch(()=>{});
}
</script>
</body>
</html>`, color, title, color, title, oid, kind, string(goalsJSON))

	manifest := fmt.Sprintf(`{
  "name": %q,
  "short_name": %q,
  "start_url": ".",
  "display": "standalone",
  "background_color": "#0b1220",
  "theme_color": %q,
  "lang": "es",
  "icons": [{"src":"icon.svg","sizes":"any","type":"image/svg+xml","purpose":"any maskable"}]
}`, title, title, color)

	sw := `self.addEventListener('install', e => { e.waitUntil(caches.open('orges-v1').then(c => c.addAll(['./','./index.html','./manifest.json','./icon.svg']))); });
self.addEventListener('fetch', e => { e.respondWith(caches.match(e.request).then(r => r || fetch(e.request))); });`

	icon := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128">
<rect width="128" height="128" rx="28" fill="%s"/>
<text x="64" y="78" text-anchor="middle" font-size="48" font-family="sans-serif" fill="#061018">A</text>
</svg>`, color)

	readme := fmt.Sprintf(`ORGES export — AlsetOS Native
Title: %s
Kind: %s
OID: %s
Target: web + PWA

Sirve esta carpeta con cualquier servidor estático HTTPS/HTTP.
En móvil: abrir index.html → "Añadir a pantalla de inicio".
Android nativo (APK) requiere empaquetado Cordova/Capacitor en fase posterior; este paquete es PWA instalable.
`, title, kind, oid)

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	files := map[string]string{
		"index.html":    html,
		"manifest.json": manifest,
		"sw.js":         sw,
		"icon.svg":      icon,
		"README.txt":    readme,
	}
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(content)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
