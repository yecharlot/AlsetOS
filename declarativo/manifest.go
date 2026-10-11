// Package declarativo implementa la rebanada vertical ATS-001:
// declaración de organismos → validación → resolución de contratos →
// ensamblaje ORGES → estado reactivo → ARL/Zyrion → capacidades → auditoría.
//
// LISPAI-AOS (subconjunto): ver lispai/aos_parse.go + declarativo/aos.go.
// JSON estructurado y .aos textual comparten OrgDecl / Runtime.
package declarativo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ContractRef es un contrato versionado (oferta o requisito).
type ContractRef struct {
	Name    string `json:"name"`
	Version string `json:"version"` // semver mayor.menor (p.ej. "1.0")
	Role    string `json:"role"`    // "provides" | "requires"
}

// Condition es una condición ARL estructurada (no parser de texto libre aún).
type Condition struct {
	Path string      `json:"path"` // átomo de estado, p.ej. "inventario.stock"
	Op   string      `json:"op"`   // eq, ne, gt, gte, lt, lte, exists, missing
	Value interface{} `json:"value,omitempty"`
}

// Rule es una regla ARL mínima: when (todas AND) → then átomo = valor.
// Conclusión Zyrion se deriva de si las condiciones son Si/No/Incierto.
type Rule struct {
	ID        string      `json:"id"`
	When      []Condition `json:"when"`
	ThenPath  string      `json:"then_path"`
	ThenValue interface{} `json:"then_value"`
	// ThenOnIndeterminate: valor a escribir si alguna condición es indeterminada
	// Si vacío, no se escribe y se registra conclusión zyrion=incierto.
	ThenOnIndeterminate interface{} `json:"then_on_indeterminate,omitempty"`
}

// OrgDecl declaración de un organismo (campos ATS-001 esenciales).
type OrgDecl struct {
	OID          string            `json:"oid"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Kind         string            `json:"kind"` // ORGES role label
	Description  string            `json:"description,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Goals        []string          `json:"goals,omitempty"`
	Contracts    []ContractRef     `json:"contracts,omitempty"`
	Rules        []Rule            `json:"rules,omitempty"`
	InitialState map[string]interface{} `json:"initial_state,omitempty"`
	// DependsOn atoms listed here are read-only from provider resolution
	Meta map[string]string `json:"meta,omitempty"`
}

// Validate comprueba campos obligatorios y contratos bien formados.
func (d OrgDecl) Validate() error {
	if strings.TrimSpace(d.OID) == "" {
		return fmt.Errorf("organismo sin oid")
	}
	if strings.TrimSpace(d.Name) == "" {
		return fmt.Errorf("organismo %s sin name", d.OID)
	}
	if strings.TrimSpace(d.Version) == "" {
		return fmt.Errorf("organismo %s sin version", d.OID)
	}
	for _, c := range d.Contracts {
		if c.Name == "" || c.Version == "" {
			return fmt.Errorf("%s: contrato incompleto", d.OID)
		}
		if c.Role != "provides" && c.Role != "requires" {
			return fmt.Errorf("%s: role de contrato inválido %q", d.OID, c.Role)
		}
	}
	for _, r := range d.Rules {
		if r.ID == "" || r.ThenPath == "" {
			return fmt.Errorf("%s: regla incompleta", d.OID)
		}
		for _, w := range r.When {
			if w.Path == "" || w.Op == "" {
				return fmt.Errorf("%s: condición incompleta en %s", d.OID, r.ID)
			}
		}
	}
	return nil
}

// DefinitionCID identidad de definición (RootCID simplificado por hash canónico).
func (d OrgDecl) DefinitionCID() (string, error) {
	if err := d.Validate(); err != nil {
		return "", err
	}
	// Canonical: sorted JSON of declaration without runtime-only fields
	canon := struct {
		OID, Name, Version, Kind, Description string
		Capabilities, Goals                   []string
		Contracts                             []ContractRef
		Rules                                 []Rule
		InitialState                          map[string]interface{}
	}{
		OID: d.OID, Name: d.Name, Version: d.Version, Kind: d.Kind,
		Description: d.Description, Capabilities: append([]string{}, d.Capabilities...),
		Goals: append([]string{}, d.Goals...), Contracts: append([]ContractRef{}, d.Contracts...),
		Rules: append([]Rule{}, d.Rules...), InitialState: d.InitialState,
	}
	sort.Strings(canon.Capabilities)
	sort.Strings(canon.Goals)
	b, err := json.Marshal(canon)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return "bafy" + hex.EncodeToString(sum[:20]), nil
}
