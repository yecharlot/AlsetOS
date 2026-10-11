// alset-plataforma: entrada orientada a REGLAS (fuente de verdad).
// El código Go es un resultado opcional de la plataforma, no el origen.
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
	case "ayuda", "help", "-h", "--help":
		uso()
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %s\n", os.Args[1])
		uso()
		os.Exit(2)
	}
}

func uso() {
	fmt.Fprintf(os.Stderr, `AlsetOS plataforma — las apps nacen de reglas

  alset-plataforma aplicar [--data DIR] [--generar DIR] <manifiesto.json>

El JSON de reglas es la fuente de verdad.
--generar emite código Go editable (artefacto), no el origen.

Ejemplo:
  alset-plataforma aplicar --generar /tmp/gen ejemplos/reglas/cafeteria.json
`)
}

func cmdAplicar(args []string) {
	var data, gen, ruta string
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
	if err != nil {
		fmt.Fprintln(os.Stderr, "runtime:", err)
		os.Exit(1)
	}
	if err := rt.AplicarReglas(m); err != nil {
		fmt.Fprintln(os.Stderr, "aplicar reglas:", err)
		os.Exit(1)
	}

	for _, o := range m.Organismos {
		if _, err := rt.EvalOrgRules(o.OID); err != nil {
			fmt.Fprintln(os.Stderr, "eval", o.OID, err)
			os.Exit(1)
		}
	}

	if gen != "" {
		if err := declarativo.GenerarCodigo(m, gen); err != nil {
			fmt.Fprintln(os.Stderr, "generar código:", err)
			os.Exit(1)
		}
		fmt.Println("código generado en", gen)
	}

	if data != "" {
		if err := rt.Persist(); err != nil {
			fmt.Fprintln(os.Stderr, "persistir:", err)
			os.Exit(1)
		}
	}

	resumen := map[string]interface{}{
		"app":        m.App,
		"version":    m.Version,
		"organismos": len(m.Organismos),
		"audit_len":  rt.AuditLen(),
		"generado":   gen,
		"mensaje":    "ensamblaje desde reglas OK — el código no fue la fuente",
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resumen)
}
