// alset-plataforma: reglas JSON o flows LISPAI-AOS.
// El código Go es resultado opcional; la fuente es reglas/.aos.
package main

import (
	"encoding/json"
	"fmt"
	"os"

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
	case "ayuda", "help", "-h", "--help":
		uso()
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %s\n", os.Args[1])
		uso()
		os.Exit(2)
	}
}

func uso() {
	fmt.Fprintf(os.Stderr, `AlsetOS plataforma — ORGES desde reglas o LISPAI-AOS

  alset-plataforma aplicar [--data DIR] [--generar DIR] <manifiesto.json>
  alset-plataforma flow    [--data DIR] [--generar DIR] <programa.aos>

LISPAI-AOS: defines orges, las unes como LEGO, la UI también es organismo.
El runtime ensambla, evalúa ARL/Zyrion y puede generar código editable.

Ejemplos:
  alset-plataforma aplicar --generar /tmp/gen ejemplos/reglas/cafeteria.json
  alset-plataforma flow    --generar /tmp/gen ejemplos/flows/mostrador.aos
`)
}

func cmdAplicar(args []string) {
	data, gen, ruta := parseFlags(args)
	if ruta == "" {
		fmt.Fprintln(os.Stderr, "falta ruta al manifiesto JSON")
		os.Exit(2)
	}
	m, err := declarativo.CargarReglas(ruta)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	rt, err := declarativo.NewRuntime(data)
	must(err)
	must(rt.AplicarReglas(m))
	for _, o := range m.Organismos {
		if _, err := rt.EvalOrgRules(o.OID); err != nil {
			fmt.Fprintln(os.Stderr, "eval", o.OID, err)
			os.Exit(1)
		}
	}
	if gen != "" {
		must(declarativo.GenerarCodigo(m, gen))
		fmt.Println("código generado en", gen)
	}
	if data != "" {
		must(rt.Persist())
	}
	printResumen(m.App, m.Version, len(m.Organismos), rt.AuditLen(), gen, "reglas-json")
}

func cmdFlow(args []string) {
	data, gen, ruta := parseFlags(args)
	if ruta == "" {
		fmt.Fprintln(os.Stderr, "falta ruta al programa .aos")
		os.Exit(2)
	}
	flow, err := declarativo.CargarAOS(ruta)
	if err != nil {
		fmt.Fprintln(os.Stderr, "AOS:", err)
		os.Exit(1)
	}
	rt, err := declarativo.NewRuntime(data)
	must(err)
	must(rt.AplicarFlow(flow))
	for _, o := range flow.Manifiesto.Organismos {
		if o.Kind == "UI" {
			continue
		}
		if _, err := rt.EvalOrgRules(o.OID); err != nil {
			fmt.Fprintln(os.Stderr, "eval", o.OID, err)
			os.Exit(1)
		}
	}
	if gen != "" {
		must(declarativo.GenerarArtefactosFlow(flow, gen))
		fmt.Println("artefactos (JSON+UI+Go) en", gen)
	}
	if data != "" {
		must(rt.Persist())
	}
	printResumen(flow.Name, flow.Version, len(flow.Manifiesto.Organismos), rt.AuditLen(), gen, "lispai-aos")
	if flow.SurfaceOID != "" {
		fmt.Fprintf(os.Stderr, "surface activa: %s\n", flow.SurfaceOID)
	}
}

func parseFlags(args []string) (data, gen, ruta string) {
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
		case len(a) > 0 && a[0] != '-':
			ruta = a
		}
	}
	return
}

func printResumen(app, ver string, nOrg, audit int, gen, src string) {
	resumen := map[string]interface{}{
		"app":        app,
		"version":    ver,
		"organismos": nOrg,
		"audit_len":  audit,
		"generado":   gen,
		"fuente":     src,
		"mensaje":    "ensamblaje OK — fuente " + src + "; código es artefacto editable",
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resumen)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

