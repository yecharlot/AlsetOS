package declarativo

import (
	"encoding/json"
	"fmt"
	"os"
)

// ManifiestoReglas es la entrada de la plataforma orientada a reglas (no a código).
// El usuario escribe esto; el código Go generado es un artefacto opcional.
type ManifiestoReglas struct {
	Plataforma  string            `json:"plataforma"` // "AlsetOS"
	App         string            `json:"app"`
	Version     string            `json:"version"`
	Descripcion string            `json:"descripcion,omitempty"`
	Organismos  []OrgDecl         `json:"organismos"`
	Meta        map[string]string `json:"meta,omitempty"`
}

// CargarReglas lee un manifiesto JSON de reglas desde archivo.
func CargarReglas(ruta string) (*ManifiestoReglas, error) {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}
	return ParsearReglas(b)
}

// ParsearReglas deserializa JSON de reglas.
func ParsearReglas(b []byte) (*ManifiestoReglas, error) {
	var m ManifiestoReglas
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("reglas JSON inválidas: %w", err)
	}
	if m.App == "" {
		return nil, fmt.Errorf("reglas: falta campo app")
	}
	if len(m.Organismos) == 0 {
		return nil, fmt.Errorf("reglas: se requiere al menos un organismo")
	}
	if m.Plataforma == "" {
		m.Plataforma = "AlsetOS"
	}
	if m.Version == "" {
		m.Version = "1.0.0"
	}
	return &m, nil
}

// AplicarReglas carga el ensamblaje en el runtime declarativo (resultado de las reglas).
func (rt *Runtime) AplicarReglas(m *ManifiestoReglas) error {
	if m == nil {
		return fmt.Errorf("manifiesto nulo")
	}
	rt.audit("reglas.aplicar", m.App, "inicio", m.Version)
	if err := rt.LoadAssembly(m.Organismos); err != nil {
		return err
	}
	rt.audit("reglas.aplicar", m.App, "ensamblaje_ok", fmt.Sprintf("%d organismos", len(m.Organismos)))
	return nil
}
