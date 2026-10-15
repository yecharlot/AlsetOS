package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yecharlot/AlsetOS/declarativo"
	"github.com/yecharlot/AlsetOS/ingest"
	"github.com/yecharlot/AlsetOS/mind"
	"github.com/yecharlot/AlsetOS/token"
)

//go:embed all:web
var embeddedStudio embed.FS


// appInstance — cada app desplegada tiene su propio Runtime (no se pisan).
type appInstance struct {
	rt   *declarativo.Runtime
	flow *declarativo.FlowAOS
}

type registry struct {
	mu   sync.Mutex
	apps map[string]*appInstance
}

func (reg *registry) put(name string, rt *declarativo.Runtime, flow *declarativo.FlowAOS) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	if reg.apps == nil {
		reg.apps = map[string]*appInstance{}
	}
	reg.apps[name] = &appInstance{rt: rt, flow: flow}
}

func (reg *registry) get(name string) *appInstance {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	if reg.apps == nil {
		return nil
	}
	return reg.apps[name]
}

func main() {
	if len(os.Args) < 2 {
		uso()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "aplicar":
		cmdAplicar(os.Args[2:])
	case "flow":
		cmdFlow(os.Args[2:])
	case "studio":
		cmdStudio(os.Args[2:])
	case "ayuda", "help", "-h", "--help":
		uso()
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %s\n", os.Args[1])
		uso()
		os.Exit(2)
	}
}

func uso() {
	fmt.Fprintf(os.Stderr, `AlsetOS plataforma — Studio ORGES multi-instancia

  alset-plataforma studio [--addr 127.0.0.1:5178] [--apps ./apps]
  Cada deploy crea /apps/{nombre}/ con Runtime propio (no se sobrescribe otra app).
`)
}

func cmdAplicar(args []string) {
	_, gen, _, ruta := parseFlags(args)
	if ruta == "" {
		fail("falta JSON")
	}
	m, err := declarativo.CargarReglas(ruta)
	must(err)
	rt, err := declarativo.NewRuntime("")
	must(err)
	must(rt.AplicarReglas(m))
	if gen != "" {
		must(declarativo.GenerarCodigo(m, gen))
	}
	printJSON(map[string]interface{}{"app": m.App, "organismos": len(m.Organismos)})
}

func cmdFlow(args []string) {
	data, gen, deploy, ruta := parseFlags(args)
	if ruta == "" {
		fail("falta .aos")
	}
	flow, err := declarativo.CargarAOS(ruta)
	must(err)
	rt, err := declarativo.NewRuntime(data)
	must(err)
	must(rt.AplicarFlow(flow))
	for _, o := range flow.Manifiesto.Organismos {
		if o.Kind != "UI" {
			_, _ = rt.EvalOrgRules(o.OID)
		}
	}
	if gen != "" {
		must(declarativo.GenerarArtefactosFlow(flow, gen))
	}
	if deploy == "" && flow.Deploy != nil {
		deploy = flow.Deploy.OutDir
	}
	if deploy != "" {
		must(declarativo.DeployFlow(flow, deploy))
	}
	printJSON(map[string]interface{}{"app": flow.Name, "deploy": deploy})
}

func cmdStudio(args []string) {
	addr := "127.0.0.1:5178"
	studioDir := "studio"
	dataDir := filepath.Join(os.TempDir(), "alset-studio-data")
	appsDir := filepath.Join(".", "apps")
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--addr":
			i++
			if i < len(args) {
				addr = args[i]
			}
		case "--studio":
			i++
			if i < len(args) {
				studioDir = args[i]
			}
		case "--data":
			i++
			if i < len(args) {
				dataDir = args[i]
			}
		case "--apps":
			i++
			if i < len(args) {
				appsDir = args[i]
			}
		}
	}
	if !filepath.IsAbs(appsDir) {
		cwd, _ := os.Getwd()
		appsDir = filepath.Join(cwd, appsDir)
	}
	_ = os.MkdirAll(appsDir, 0o755)
	reg := &registry{apps: map[string]*appInstance{}}

	bootApp := func(name, src string) (*declarativo.FlowAOS, *declarativo.Runtime, error) {
		flow, err := declarativo.ParsearAOS(src)
		if err != nil {
			return nil, nil, err
		}
		if name != "" {
			flow.Name = sanitizeApp(name)
			flow.Manifiesto.App = flow.Name
		} else {
			flow.Name = sanitizeApp(flow.Name)
			flow.Manifiesto.App = flow.Name
		}
		rtDir := filepath.Join(dataDir, "rt", flow.Name)
		_ = os.MkdirAll(rtDir, 0o755)
		rt, err := declarativo.NewRuntime(rtDir)
		if err != nil {
			return nil, nil, err
		}
		if err := rt.AplicarFlow(flow); err != nil {
			return nil, nil, err
		}
		for _, o := range flow.Manifiesto.Organismos {
			if o.Kind != "UI" {
				_, _ = rt.EvalOrgRules(o.OID)
			}
		}
		reg.put(flow.Name, rt, flow)
		return flow, rt, nil
	}

	mux := http.NewServeMux()
	// Studio UI: directorio local si existe; si no, embebido (exe Windows portable)
	if st, err := os.Stat(studioDir); err == nil && st.IsDir() {
		mux.Handle("/", http.FileServer(http.Dir(studioDir)))
	} else {
		sub, err := fs.Sub(embeddedStudio, "web")
		if err != nil {
			fmt.Fprintln(os.Stderr, "embed studio:", err)
			os.Exit(1)
		}
		mux.Handle("/", http.FileServer(http.FS(sub)))
		fmt.Fprintln(os.Stderr, "  UI: embebida en el ejecutable")
	}
	// /apps/{name}/api/action|state — mismo contrato que /api/apps/{name}/…
	// (app.html usa rutas relativas bajo /apps/nombre/)
	appsFS := http.StripPrefix("/apps/", http.FileServer(http.Dir(appsDir)))
	mux.HandleFunc("/apps/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/apps/")
		parts := strings.Split(path, "/")
		if len(parts) >= 3 && parts[1] == "api" {
			name := sanitizeApp(parts[0])
			action := parts[2]
			inst := reg.get(name)
			if inst == nil {
				srcPath := filepath.Join(appsDir, name, "source.aos")
				if b, err := os.ReadFile(srcPath); err == nil {
					if _, _, err := bootApp(name, string(b)); err == nil {
						inst = reg.get(name)
					}
				}
			}
			switch action {
			case "state":
				st := map[string]interface{}{}
				if inst != nil && inst.rt != nil && inst.rt.State != nil {
					st = inst.rt.State.Snapshot()
				}
				writeJSON(w, map[string]interface{}{"ok": true, "app": name, "state": st})
				return
			case "action":
				if r.Method != http.MethodPost {
					http.Error(w, "POST", 405)
					return
				}
				if inst == nil {
					writeErr(w, fmt.Errorf("app %s no cargada — vuelve a Desplegar", name))
					return
				}
				body := readBody(r)
				res, err := inst.rt.HandleEmit(body.Emit)
				if err != nil {
					writeErr(w, err)
					return
				}
				writeJSON(w, map[string]interface{}{"ok": true, "app": name, "state": res["state"], "changed": res["changed"]})
				return
			}
		}
		appsFS.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/help", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"product": "Alset Studio ORGES multi-instancia",
			"note":    "Cada app en /apps/{nombre}/ tiene Runtime y estado propios",
		})
	})

	mindAgent := mind.NewAgent()
	tokenReg, errTok := token.NewRegistry(filepath.Join(appsDir, "..", "data", "tokens"))
	if errTok != nil {
		tokenReg, _ = token.NewRegistry("data/tokens")
	}
	mux.HandleFunc("/api/mind/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, fmt.Errorf("POST required"))
			return
		}
		var req mind.AnalyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, mindAgent.Handle(req))
	})
	mux.HandleFunc("/api/mind/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, fmt.Errorf("POST required"))
			return
		}
		var req mind.AnalyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, err)
			return
		}
		req.Mode = "chat"
		writeJSON(w, mindAgent.Handle(req))
	})

	
	mux.HandleFunc("/api/work/modes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{"ok": true, "modes": mind.CatalogModes()})
	})
	mux.HandleFunc("/api/work/domain", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		var req mind.AnalyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, err)
			return
		}
		if req.WorkMode == "" {
			req.WorkMode = req.Mode
		}
		if req.WorkMode == "" {
			req.WorkMode = "analyze"
		}
		writeJSON(w, mindAgent.Handle(req))
	})
	mux.HandleFunc("/api/ingest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		var req mind.IngestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, mindAgent.Ingest(req))
	})

	mux.HandleFunc("/api/finance/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			writeErr(w, err)
			return
		}
		f, hdr, err := r.FormFile("file")
		if err != nil {
			writeErr(w, fmt.Errorf("file requerido: %w", err))
			return
		}
		defer f.Close()
		tmp, err := os.CreateTemp("", "alset-fin-*"+filepath.Ext(hdr.Filename))
		if err != nil {
			writeErr(w, err)
			return
		}
		tmpPath := tmp.Name()
		defer os.Remove(tmpPath)
		if _, err := io.Copy(tmp, f); err != nil {
			tmp.Close()
			writeErr(w, err)
			return
		}
		tmp.Close()
		res, err := ingest.ExtractFromFile(tmpPath)
		if res == nil {
			writeErr(w, err)
			return
		}
		name := r.FormValue("name")
		if name == "" {
			name = strings.TrimSuffix(filepath.Base(hdr.Filename), filepath.Ext(hdr.Filename))
		}
		payload, _ := json.Marshal(res.Statement)
		ing := mindAgent.Ingest(mind.IngestRequest{
			Kind:    "finance-json",
			Name:    name,
			Payload: payload,
		})
		if len(res.Warnings) > 0 {
			ing.Warnings = append(res.Warnings, ing.Warnings...)
		}
		if err != nil {
			ing.Warnings = append(ing.Warnings, err.Error())
		}
		writeJSON(w, ing)
	})

	mux.HandleFunc("/api/architect/interview", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		var in mind.InterviewInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, mindAgent.AnalyzeInterview(in))
	})
	mux.HandleFunc("/api/token/mint", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		var body struct {
			Kind    string            `json:"kind"`
			Name    string            `json:"name"`
			Desc    string            `json:"description"`
			Version string            `json:"version"`
			Domain  string            `json:"domain"`
			Source  string            `json:"source"`
			Meta    map[string]string `json:"meta"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, err)
			return
		}
		if body.Source == "" {
			writeErr(w, fmt.Errorf("source .aos requerido"))
			return
		}
		tok, ev, err := tokenReg.Mint(body.Kind, body.Name, body.Desc, body.Version, body.Domain, body.Source, body.Meta)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, map[string]interface{}{"ok": true, "token": tok, "event": ev})
	})
	mux.HandleFunc("/api/token/list", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{"ok": true, "tokens": tokenReg.List()})
	})
	mux.HandleFunc("/api/token/append", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		var body struct {
			ID      string         `json:"id"`
			Type    string         `json:"type"`
			Payload map[string]any `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, err)
			return
		}
		ev, err := tokenReg.Append(body.ID, body.Type, body.Payload)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, map[string]interface{}{"ok": true, "event": ev})
	})
	mux.HandleFunc("/api/apps", func(w http.ResponseWriter, r *http.Request) {
		entries, _ := os.ReadDir(appsDir)
		var names []string
		for _, e := range entries {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
		reg.mu.Lock()
		live := make([]string, 0, len(reg.apps))
		for k := range reg.apps {
			live = append(live, k)
		}
		reg.mu.Unlock()
		writeJSON(w, map[string]interface{}{"apps": names, "live": live})
	})

	mux.HandleFunc("/api/flow/apply", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(r)
		flow, rt, err := bootApp(body.App, body.Source)
		if err != nil {
			writeErr(w, err)
			return
		}
		st := rt.State.Snapshot()
		writeJSON(w, map[string]interface{}{
			"ok": true, "app": flow.Name, "surface": flow.SurfaceOID,
			"surfaces": flow.Surfaces, "organismos": len(flow.Manifiesto.Organismos),
			"state": st,
		})
	})

	mux.HandleFunc("/api/flow/test", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(r)
		flow, err := declarativo.ParsearAOS(body.Source)
		if err != nil {
			writeErr(w, err)
			return
		}
		name := sanitizeApp(body.App)
		if name == "" {
			name = sanitizeApp(flow.Name)
		}
		inst := reg.get(name)
		var rt *declarativo.Runtime
		if inst != nil {
			rt = inst.rt
			flow = inst.flow
		} else {
			_, rt2, err := bootApp(name, body.Source)
			if err != nil {
				writeErr(w, err)
				return
			}
			rt = rt2
		}
		errs := rt.RunFlowTests(flow)
		writeJSON(w, map[string]interface{}{"ok": len(errs) == 0, "errors": errs, "app": name})
	})

	mux.HandleFunc("/api/flow/action", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(r)
		name := sanitizeApp(body.App)
		inst := reg.get(name)
		if inst == nil {
			writeErr(w, fmt.Errorf("app %q no está activa — Aplicar o Desplegar esa plantilla primero", name))
			return
		}
		res, err := inst.rt.HandleEmit(body.Emit)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, map[string]interface{}{"ok": true, "app": name, "changed": res["changed"], "state": res["state"]})
	})

	mux.HandleFunc("/api/flow/deploy", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(r)
		flow, rt, err := bootApp(body.App, body.Source)
		if err != nil {
			writeErr(w, err)
			return
		}
		// Instancia dedicada: carpeta propia por nombre de app
		dir, err := declarativo.DeployProduction(flow, appsDir)
		if err != nil {
			writeErr(w, err)
			return
		}
		_ = rt.Persist()
		// Persist source for cold restart
		_ = os.WriteFile(filepath.Join(dir, "source.aos"), []byte(body.Source), 0o644)
		if body.Layouts != nil {
			lb, _ := json.MarshalIndent(body.Layouts, "", "  ")
			_ = os.WriteFile(filepath.Join(dir, "layout.json"), lb, 0o644)
		}
		writeJSON(w, map[string]interface{}{
			"ok": true, "app": flow.Name, "out": dir,
			"url": "/apps/" + flow.Name + "/", "production": true,
			"isolated": true,
		})
	})

	mux.HandleFunc("/api/apps/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/apps/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 || parts[0] == "" {
			http.NotFound(w, r)
			return
		}
		name := sanitizeApp(parts[0])
		action := ""
		if len(parts) > 1 {
			action = parts[1]
		}
		inst := reg.get(name)
		// Cold load from source.aos if not live
		if inst == nil {
			srcPath := filepath.Join(appsDir, name, "source.aos")
			if b, err := os.ReadFile(srcPath); err == nil {
				if _, _, err := bootApp(name, string(b)); err == nil {
					inst = reg.get(name)
				}
			}
		}
		switch action {
		case "state":
			st := map[string]interface{}{}
			if inst != nil && inst.rt != nil && inst.rt.State != nil {
				st = inst.rt.State.Snapshot()
			} else if b, err := os.ReadFile(filepath.Join(appsDir, name, "initial_state.json")); err == nil {
				_ = json.Unmarshal(b, &st)
			}
			writeJSON(w, map[string]interface{}{"ok": true, "app": name, "state": st})
		case "action":
			if r.Method != http.MethodPost {
				http.Error(w, "POST", 405)
				return
			}
			if inst == nil {
				writeErr(w, fmt.Errorf("app %s no cargada", name))
				return
			}
			body := readBody(r)
			res, err := inst.rt.HandleEmit(body.Emit)
			if err != nil {
				writeErr(w, err)
				return
			}
			writeJSON(w, map[string]interface{}{"ok": true, "app": name, "state": res["state"], "changed": res["changed"]})
		default:
			http.NotFound(w, r)
		}
	})

	fmt.Fprintf(os.Stderr, "Alset Studio ORGES → http://%s/\n", addr)
	fmt.Fprintf(os.Stderr, "  Apps aisladas → %s  (URL /apps/{nombre}/)\n", appsDir)
	must(http.ListenAndServe(addr, mux))
}

func sanitizeApp(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == ' ' {
			b.WriteByte('-')
		}
	}
	out := b.String()
	if out == "" {
		return "app"
	}
	return out
}

type reqBody struct {
	Source     string                 `json:"source"`
	App        string                 `json:"app"`
	Emit       string                 `json:"emit"`
	Production bool                   `json:"production"`
	Layouts    map[string]interface{} `json:"layouts"`
}

func readBody(r *http.Request) reqBody {
	var b reqBody
	_ = json.NewDecoder(r.Body).Decode(&b)
	return b
}

func parseFlags(args []string) (data, gen, deploy, ruta string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--data" || a == "-data":
			i++
			if i < len(args) {
				data = args[i]
			}
		case a == "--generar" || a == "-generar":
			i++
			if i < len(args) {
				gen = args[i]
			}
		case a == "--deploy" || a == "-deploy":
			i++
			if i < len(args) {
				deploy = args[i]
			}
		case len(a) > 0 && a[0] != '-':
			ruta = a
		}
	}
	return
}

func printJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, err error) {
	w.WriteHeader(400)
	writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func fail(msg string) { fmt.Fprintln(os.Stderr, msg); os.Exit(2) }
