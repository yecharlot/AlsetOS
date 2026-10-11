package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yecharlot/AlsetOS/declarativo"
)

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
	fmt.Fprintf(os.Stderr, `AlsetOS plataforma — LISPAI-AOS + Studio

  alset-plataforma aplicar [--data DIR] [--generar DIR] <manifiesto.json>
  alset-plataforma flow    [--data DIR] [--generar DIR] [--deploy DIR] <programa.aos>
  alset-plataforma studio  [--addr 127.0.0.1:5178] [--studio DIR]

Studio: IDE visual LEGO + preview mobile/tablet/desktop + deploy desde flow.
`)
}

func cmdAplicar(args []string) {
	data, gen, deploy, ruta := parseFlags(args)
	_ = deploy
	if ruta == "" {
		fail("falta manifiesto JSON")
	}
	m, err := declarativo.CargarReglas(ruta)
	must(err)
	rt, err := declarativo.NewRuntime(data)
	must(err)
	must(rt.AplicarReglas(m))
	for _, o := range m.Organismos {
		_, _ = rt.EvalOrgRules(o.OID)
	}
	if gen != "" {
		must(declarativo.GenerarCodigo(m, gen))
	}
	if data != "" {
		_ = rt.Persist()
	}
	printJSON(map[string]interface{}{"app": m.App, "organismos": len(m.Organismos), "fuente": "json"})
}

func cmdFlow(args []string) {
	data, gen, deploy, ruta := parseFlags(args)
	if ruta == "" {
		fail("falta programa .aos")
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
	testErrs := rt.RunFlowTests(flow)
	if gen != "" {
		must(declarativo.GenerarArtefactosFlow(flow, gen))
	}
	outDeploy := deploy
	if outDeploy == "" && flow.Deploy != nil {
		outDeploy = flow.Deploy.OutDir
	}
	if outDeploy != "" {
		must(declarativo.DeployFlow(flow, outDeploy))
		fmt.Fprintln(os.Stderr, "deploy pack:", outDeploy)
	}
	if data != "" {
		_ = rt.Persist()
	}
	printJSON(map[string]interface{}{
		"app": flow.Name, "organismos": len(flow.Manifiesto.Organismos),
		"surface": flow.SurfaceOID, "tests_fail": testErrs, "fuente": "lispai-aos",
		"deploy": outDeploy,
	})
}

func cmdStudio(args []string) {
	addr := "127.0.0.1:5178"
	studioDir := "studio"
	dataDir := filepath.Join(os.TempDir(), "alset-studio-data")
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--addr":
			i++; if i < len(args) { addr = args[i] }
		case "--studio":
			i++; if i < len(args) { studioDir = args[i] }
		case "--data":
			i++; if i < len(args) { dataDir = args[i] }
		}
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(studioDir)))
	mux.HandleFunc("/api/flow/apply", func(w http.ResponseWriter, r *http.Request) {
		src := readSource(r)
		flow, err := declarativo.ParsearAOS(src)
		if err != nil {
			writeErr(w, err); return
		}
		rt, err := declarativo.NewRuntime(dataDir)
		if err != nil {
			writeErr(w, err); return
		}
		if err := rt.AplicarFlow(flow); err != nil {
			writeErr(w, err); return
		}
		writeJSON(w, map[string]interface{}{
			"ok": true, "app": flow.Name, "surface": flow.SurfaceOID,
			"surfaces": flow.Surfaces, "organismos": len(flow.Manifiesto.Organismos),
		})
	})
	mux.HandleFunc("/api/flow/test", func(w http.ResponseWriter, r *http.Request) {
		src := readSource(r)
		flow, err := declarativo.ParsearAOS(src)
		if err != nil {
			writeErr(w, err); return
		}
		rt, err := declarativo.NewRuntime(dataDir)
		if err != nil {
			writeErr(w, err); return
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
	mux.HandleFunc("/api/flow/deploy", func(w http.ResponseWriter, r *http.Request) {
		src := readSource(r)
		flow, err := declarativo.ParsearAOS(src)
		if err != nil {
			writeErr(w, err); return
		}
		out := filepath.Join(dataDir, "dist", flow.Name)
		if err := declarativo.DeployFlow(flow, out); err != nil {
			writeErr(w, err); return
		}
		writeJSON(w, map[string]interface{}{"ok": true, "out": out, "url": "/dist/" + flow.Name + "/"})
	})
	fmt.Fprintf(os.Stderr, "Alset Studio LISPAI-AOS → http://%s/\n", addr)
	must(http.ListenAndServe(addr, mux))
}

func readSource(r *http.Request) string {
	var body struct {
		Source string `json:"source"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	return body.Source
}

func parseFlags(args []string) (data, gen, deploy, ruta string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--data" || a == "-data":
			i++; if i < len(args) { data = args[i] }
		case a == "--generar" || a == "-generar":
			i++; if i < len(args) { gen = args[i] }
		case a == "--deploy" || a == "-deploy":
			i++; if i < len(args) { deploy = args[i] }
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
