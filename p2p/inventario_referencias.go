package p2p

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ReferenciaContenido resume las referencias conocidas de un RootCID.
// CandidatoSinReferencia es informativo: nunca autoriza borrar el bloque.
type ReferenciaContenido struct {
	RootCID string `json:"root_cid"`
	Versiones int `json:"versiones"`
	Retenciones int `json:"retenciones"`
	LeasesActivos int `json:"leases_activos"`
	OrganismosLegados int `json:"organismos_legados"`
	Presente bool `json:"presente"`
	Integro bool `json:"integro"`
	Referenciado bool `json:"referenciado"`
	CandidatoSinReferencia bool `json:"candidato_sin_referencia"`
}

func (a *AlmacenContenido) Raices() ([]string, error) {
	if a == nil { return nil, fmt.Errorf("almacén nulo") }
	a.mu.RLock()
	defer a.mu.RUnlock()
	resultado := make([]string, 0)
	if a.dir == "" {
		for root := range a.memoria { resultado = append(resultado, root) }
	} else {
		shards, err := os.ReadDir(a.dir)
		if err != nil { return nil, fmt.Errorf("listar shards de contenido: %w", err) }
		for _, shard := range shards {
			if !shard.IsDir() || len(shard.Name()) != 2 { continue }
			archivos, err := os.ReadDir(filepath.Join(a.dir, shard.Name()))
			if err != nil { return nil, fmt.Errorf("listar shard %s: %w", shard.Name(), err) }
			for _, archivo := range archivos {
				if archivo.IsDir() || strings.HasPrefix(archivo.Name(), ".") { continue }
				root := "rootcid:" + archivo.Name()
				digest, err := rutaBloque(root)
				if err != nil || !strings.HasPrefix(digest, shard.Name()) { continue }
				resultado = append(resultado, root)
			}
		}
	}
	sort.Strings(resultado)
	return resultado, nil
}

// InventarioReferencias cruza bloques locales con versiones, pins, leases
// vigentes y el índice legado. Los bloques sin referencias quedan marcados
// como candidatos de revisión manual, no como datos eliminables.
func (n *Nodo) InventarioReferencias() ([]ReferenciaContenido, error) {
	if n == nil || n.almacen == nil || n.versiones == nil { return nil, fmt.Errorf("servicios de inventario no inicializados") }
	porRoot := map[string]*ReferenciaContenido{}
	obtener := func(root string) *ReferenciaContenido {
		if porRoot[root] == nil { porRoot[root] = &ReferenciaContenido{RootCID:root} }
		return porRoot[root]
	}
	roots, err := n.almacen.Raices()
	if err != nil { return nil, err }
	for _, root := range roots { obtener(root) }
	for _, d := range n.versiones.Todas() { obtener(d.RootCID).Versiones++ }
	if n.retenciones != nil { for _, r := range n.retenciones.Todas() { obtener(r.Descriptor.RootCID).Retenciones++ } }
	if n.leases != nil { for _, l := range n.leases.Activos(time.Now().UTC()) { obtener(l.RootCID).LeasesActivos++ } }
	n.mu.RLock()
	for root := range n.organismos { obtener(root).OrganismosLegados++ }
	n.mu.RUnlock()
	resultado := make([]ReferenciaContenido, 0, len(porRoot))
	for _, item := range porRoot {
		contenido, err := n.almacen.Obtener(item.RootCID)
		item.Presente = err == nil || !errors.Is(err, ErrContenidoNoDisponible)
		item.Integro = err == nil
		_ = contenido
		item.Referenciado = item.Versiones+item.Retenciones+item.LeasesActivos+item.OrganismosLegados > 0
		item.CandidatoSinReferencia = item.Presente && !item.Referenciado
		resultado = append(resultado, *item)
	}
	sort.Slice(resultado, func(i,j int) bool { return resultado[i].RootCID < resultado[j].RootCID })
	return resultado, nil
}
