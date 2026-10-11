package declarativo

import (
	"fmt"
	"strconv"
	"strings"
)

// Binding registro de por qué se eligió un proveedor.
type Binding struct {
	ConsumerOID string `json:"consumer_oid"`
	Contract    string `json:"contract"`
	Version     string `json:"version"`
	ProviderOID string `json:"provider_oid"`
	Reason      string `json:"reason"`
}

// Assembly grafo resuelto de ORGES.
type Assembly struct {
	Orgs     map[string]OrgDecl
	Bindings []Binding
	// ProviderFor[consumerOID+"|"+contract] = providerOID
	ProviderFor map[string]string
}

func majorMinor(v string) (int, int, error) {
	parts := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if len(parts) < 1 {
		return 0, 0, fmt.Errorf("versión inválida %q", v)
	}
	maj, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	min := 0
	if len(parts) > 1 {
		min, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, err
		}
	}
	return maj, min, nil
}

// Compatible: mismo major y provider.minor >= required.minor
func versionCompatible(required, provided string) bool {
	rm, rn, err1 := majorMinor(required)
	pm, pn, err2 := majorMinor(provided)
	if err1 != nil || err2 != nil {
		return required == provided
	}
	return pm == rm && pn >= rn
}

// Resolve ensambla organismos: cada requires debe tener exactamente un provides compatible.
// Rechaza ambigüedad (varios providers) y ausencia.
func Resolve(decls []OrgDecl) (*Assembly, error) {
	byOID := map[string]OrgDecl{}
	providers := map[string][]OrgDecl{} // contract name → orgs that provide

	for _, d := range decls {
		if err := d.Validate(); err != nil {
			return nil, err
		}
		if _, exists := byOID[d.OID]; exists {
			return nil, fmt.Errorf("OID duplicado: %s", d.OID)
		}
		byOID[d.OID] = d
		for _, c := range d.Contracts {
			if c.Role == "provides" {
				providers[c.Name] = append(providers[c.Name], d)
			}
		}
	}

	a := &Assembly{
		Orgs:        byOID,
		ProviderFor: map[string]string{},
	}

	for _, d := range decls {
		for _, c := range d.Contracts {
			if c.Role != "requires" {
				continue
			}
			cands := providers[c.Name]
			var compatible []OrgDecl
			for _, p := range cands {
				for _, pc := range p.Contracts {
					if pc.Role == "provides" && pc.Name == c.Name && versionCompatible(c.Version, pc.Version) {
						compatible = append(compatible, p)
						break
					}
				}
			}
			if len(compatible) == 0 {
				return nil, fmt.Errorf("dependencia irresoluble: %s requiere %s@%s (sin proveedor compatible)", d.OID, c.Name, c.Version)
			}
			if len(compatible) > 1 {
				ids := []string{}
				for _, p := range compatible {
					ids = append(ids, p.OID)
				}
				return nil, fmt.Errorf("acoplamiento ambiguo: %s requiere %s@%s; candidatos %v — especifique un único proveedor", d.OID, c.Name, c.Version, ids)
			}
			prov := compatible[0]
			key := d.OID + "|" + c.Name
			a.ProviderFor[key] = prov.OID
			a.Bindings = append(a.Bindings, Binding{
				ConsumerOID: d.OID,
				Contract:    c.Name,
				Version:     c.Version,
				ProviderOID: prov.OID,
				Reason:      fmt.Sprintf("único provider compatible %s@%s", prov.OID, c.Version),
			})
		}
	}

	// Detect simple require cycles via provider graph
	if err := detectCycles(a); err != nil {
		return nil, err
	}
	return a, nil
}

func detectCycles(a *Assembly) error {
	// edge consumer -> provider for each binding
	adj := map[string][]string{}
	for _, b := range a.Bindings {
		adj[b.ConsumerOID] = append(adj[b.ConsumerOID], b.ProviderOID)
	}
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[string]int{}
	var visit func(string) error
	visit = func(n string) error {
		color[n] = gray
		for _, m := range adj[n] {
			switch color[m] {
			case gray:
				return fmt.Errorf("ciclo de dependencias detectado en %s → %s", n, m)
			case white:
				if err := visit(m); err != nil {
					return err
				}
			}
		}
		color[n] = black
		return nil
	}
	for oid := range a.Orgs {
		if color[oid] == white {
			if err := visit(oid); err != nil {
				return err
			}
		}
	}
	return nil
}
