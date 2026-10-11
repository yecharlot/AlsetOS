package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yecharlot/AlsetOS/declarativo"
)

type session struct {
	mu    sync.Mutex
	rt    *declarativo.Runtime
	flow  *declarativo.FlowAOS
	app   string
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
	fmt.Fprintf(os.Stderr, `AlsetOS plataforma — Studio ORGES (estilo Alset Studio)

  alset-plataforma studio [--addr 127.0.0.1:5178] [--studio studio] [--apps ./apps]
  alset-plataforma flow --deploy DIR programa.aos
  alset-plataforma aplicar manifiesto.json
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
	printJSON(map[string]interface{}{"app": flow.Name, "deploy": deploy, "tests": rt.RunFlowTests(flow)})
}

func cmdStudio(args []string) {
	addr := "127.0.0.1:5178"
	studioDir := "studio"
	dataDir := filepath.Join(os.TempDir(), "alset-studio-data")
	appsDir := filepath.Join(dataDir, "apps")
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
				appsDir = filepath.Join(dataDir, "apps")
			}
		case "--apps":
			i++
			if i < len(args) {
				appsDir = args[i]
			}
		}
	}
	_ = os.MkdirAll(appsDir, 0o755)
	sess := &session{}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(studioDir)))
	mux.Handle("/apps/", http.StripPrefix("/apps/", http.FileServer(http.Dir(appsDir))))

	mux.HandleFunc("/api/help", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"product": "Alset Studio ORGES",
			"forms":   []string{"flow", "orge", "provides", "requires", "bind", "state", "rule", "ui", "assemble", "surface", "test", "deploy"},
			"emits":   []string{"stock.inc", "stock.dec", "venta.registrar", "campo=valor"},
			"deploy":  "POST /api/flow/deploy → /apps/{name}/ producción",
		})
	})

	mux.HandleFunc("/api/apps", func(w http.ResponseWriter, r *http.Request) {
		entries, _ := os.ReadDir(appsDir)
		var names []string
		for _, e := range entries {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
		writeJSON(w, map[string]interface{}{"apps": names})
	})

	mux.HandleFunc("/api/flow/apply", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(r)
		src := body.Source
		if src == "" {
			writeErr(w, fmt.Errorf("source vacío"))
			return
		}
		flow, err := declarativo.ParsearAOS(src)
		if err != nil {
			writeErr(w, err)
			return
		}
		if body.App != "" {
			flow.Name = body.App
			flow.Manifiesto.App = body.App
		}
		rt, err := declarativo.NewRuntime(filepath.Join(dataDir, "rt", flow.Name))
		if err != nil {
			writeErr(w, err)
			return
		}
		if err := rt.AplicarFlow(flow); err != nil {
			writeErr(w, err)
			return
		}
		for _, o := range flow.Manifiesto.Organismos {
			if o.Kind != "UI" {
				_, _ = rt.EvalOrgRules(o.OID)
			}
		}
		sess.mu.Lock()
		sess.rt, sess.flow, sess.app = rt, flow, flow.Name
		sess.mu.Unlock()
		st := map[string]interface{}{}
		if rt.State != nil {
			st = rt.State.Snapshot()
		}
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
		rt, err := declarativo.NewRuntime("")
		if err != nil {
			writeErr(w, err)
			return
		}
		_ = rt.AplicarFlow(flow)
		for _, o := range flow.Manifiesto.Organismos {
			if o.Kind != "UI" {
				_, _ = rt.EvalOrgRules(o.OID)
			}
		}
		errs := rt.RunFlowTests(flow)
		writeJSON(w, map[string]interface{}{"ok": len(errs) == 0, "errors": errs})
	})

	mux.HandleFunc("/api/flow/action", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(r)
		sess.mu.Lock()
		rt, flow := sess.rt, sess.flow
		sess.mu.Unlock()
		if rt == nil || flow == nil {
			writeErr(w, fmt.Errorf("aplica el flow primero"))
			return
		}
		res, err := rt.HandleEmit(body.Emit)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, map[string]interface{}{"ok": true, "changed": res["changed"], "state": res["state"]})
	})

	mux.HandleFunc("/api/flow/deploy", func(w http.ResponseWriter, r *http.Request) {
		body := readBody(r)
		flow, err := declarativo.ParsearAOS(body.Source)
		if err != nil {
			writeErr(w, err)
			return
		}
		if body.App != "" {
			flow.Name = body.App
			flow.Manifiesto.App = body.App
		}
		rt, err := declarativo.NewRuntime(filepath.Join(dataDir, "rt", flow.Name))
		if err != nil {
			writeErr(w, err)
			return
		}
		if err := rt.AplicarFlow(flow); err != nil {
			writeErr(w, err)
			return
		}
		for _, o := range flow.Manifiesto.Organismos {
			if o.Kind != "UI" {
				_, _ = rt.EvalOrgRules(o.OID)
			}
		}
		dir, err := declarativo.DeployProduction(flow, appsDir)
		if err != nil {
			writeErr(w, err)
			return
		}
		_ = rt.Persist()
		sess.mu.Lock()
		sess.rt, sess.flow, sess.app = rt, flow, flow.Name
		sess.mu.Unlock()
		writeJSON(w, map[string]interface{}{
			"ok": true, "out": dir, "url": "/apps/" + flow.Name + "/",
			"production": true,
		})
	})

	// Per-app live API for production shells
	mux.HandleFunc("/api/apps/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/apps/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 || parts[0] == "" {
			http.NotFound(w, r)
			return
		}
		name := parts[0]
		action := ""
		if len(parts) > 1 {
			action = parts[1]
		}
		sess.mu.Lock()
		rt := sess.rt
		app := sess.app
		sess.mu.Unlock()
		switch action {
		case "state":
			st := map[string]interface{}{}
			if rt != nil && app == name && rt.State != nil {
				st = rt.State.Snapshot()
			} else {
				// load initial_state.json
				b, err := os.ReadFile(filepath.Join(appsDir, name, "initial_state.json"))
				if err == nil {
					_ = json.Unmarshal(b, &st)
				}
			}
			writeJSON(w, map[string]interface{}{"ok": true, "state": st})
		case "action":
			if r.Method != http.MethodPost {
				http.Error(w, "POST", 405)
				return
			}
			body := readBody(r)
			if rt == nil || app != name {
				writeErr(w, fmt.Errorf("app %s no activa en sesión — abrir desde Studio tras Desplegar o Aplicar", name))
				return
			}
			res, err := rt.HandleEmit(body.Emit)
			if err != nil {
				writeErr(w, err)
				return
			}
			writeJSON(w, map[string]interface{}{"ok": true, "state": res["state"], "changed": res["changed"]})
		default:
			http.NotFound(w, r)
		}
	})

	fmt.Fprintf(os.Stderr, "Alset Studio ORGES → http://%s/\n", addr)
	fmt.Fprintf(os.Stderr, "  Apps producción → http://%s/apps/\n", addr)
	must(http.ListenAndServe(addr, mux))
}

type reqBody struct {
	Source     string `json:"source"`
	App        string `json:"app"`
	Emit       string `json:"emit"`
	Production bool   `json:"production"`
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
