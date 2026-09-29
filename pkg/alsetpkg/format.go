// Package alsetpkg defines the .alset package format (Alset-native, not TCZ).
package alsetpkg

import (
	"encoding/json"
	"os"
	"time"
)

// Manifest is the root of a .alset package (JSON).
// Extension: *.alset or *.alset.json
type Manifest struct {
	Format      string            `json:"format"` // alset-pkg/v1 or alset-app/v1
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Kind        string            `json:"kind"` // app | desktop | organism | shell
	Description string            `json:"description,omitempty"`
	Created     string            `json:"created,omitempty"`
	Requires    []string          `json:"requires,omitempty"` // other package names
	Entrypoint  string            `json:"entrypoint,omitempty"`
	Files       map[string]string `json:"files,omitempty"` // path -> relative content path or inline
	Tree        json.RawMessage   `json:"tree,omitempty"`  // UI tree for apps
	Organism    json.RawMessage   `json:"organism,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
}

func NewApp(name, version string) Manifest {
	return Manifest{
		Format:  "alset-app/v1",
		Name:    name,
		Version: version,
		Kind:    "app",
		Created: time.Now().UTC().Format(time.RFC3339),
	}
}

func Load(path string) (Manifest, error) {
	var m Manifest
	b, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(b, &m)
	return m, err
}

func Save(path string, m Manifest) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
