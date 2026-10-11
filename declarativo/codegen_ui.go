package declarativo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GenerarArtefactosFlow escribe JSON de reglas + IR UI + Go generado (editables).
// Fuente canónica sigue siendo el .aos.
func GenerarArtefactosFlow(flow *FlowAOS, dir string) error {
	if flow == nil {
		return fmt.Errorf("flow nulo")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// 1) manifiesto JSON (equivalente a reglas)
	jb, err := json.MarshalIndent(flow.Manifiesto, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "reglas.generated.json"), jb, 0o644); err != nil {
		return err
	}
	// 2) superficies UI
	ub, err := json.MarshalIndent(flow.Surfaces, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "ui.surfaces.json"), ub, 0o644); err != nil {
		return err
	}
	// 3) Go organismos
	if err := GenerarCodigo(flow.Manifiesto, dir); err != nil {
		return err
	}
	readme := fmt.Sprintf(`# Artefactos generados — %s

Fuente canónica: programa **LISPAI-AOS** (archivo .aos).

| Archivo | Uso |
|---------|-----|
| reglas.generated.json | Manifiesto de organismos (editable) |
| ui.surfaces.json | IR de UI (editable) |
| organismos_gen.go | Declaraciones Go (editable) |

Al re-aplicar el .aos se regeneran. Puedes retocar estos archivos a mano;
si re-generas, guarda tus cambios fuera o haz merge.

Surface activa: %s
`, flow.Name, flow.SurfaceOID)
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644)
}
