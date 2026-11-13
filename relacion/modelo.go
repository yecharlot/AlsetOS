// Package relacion provides a dependency-free relational model for AlsetOS.
// It does not infer causality, consciousness, or a numeric abstraction hierarchy.
package relacion

import (
 "fmt"
 "sort"
 "strings"
 "sync"
)

type Evidencia string
const (
 EvidenciaDesconocida Evidencia = "desconocida"
 EvidenciaObservada Evidencia = "observada"
 EvidenciaDocumentada Evidencia = "documentada"
 EvidenciaHipotesis Evidencia = "hipotesis"
 EvidenciaValidada Evidencia = "validada"
)

type Nodo struct {
 OID string
 Tipo string
 Meta map[string]string
}

type Relacion struct {
 ID string
 Origen string
 Destino string
 Tipo string
 Direccion string
 Temporalidad string
 Evidencia Evidencia
 Condiciones []string
 Meta map[string]string
}

func (r Relacion) Validar() error {
 if strings.TrimSpace(r.ID)=="" { return fmt.Errorf("relación sin id") }
 if strings.TrimSpace(r.Origen)=="" || strings.TrimSpace(r.Destino)=="" { return fmt.Errorf("relación %s: origen y destino obligatorios",r.ID) }
 if strings.TrimSpace(r.Tipo)=="" { return fmt.Errorf("relación %s: tipo obligatorio",r.ID) }
 switch r.Evidencia {
 case "", EvidenciaDesconocida, EvidenciaObservada, EvidenciaDocumentada, EvidenciaHipotesis, EvidenciaValidada:
 default: return fmt.Errorf("relación %s: evidencia desconocida %q",r.ID,r.Evidencia)
 }
 return nil
}

type Grafo struct {
 mu sync.RWMutex
 nodos map[string]Nodo
 relaciones map[string]Relacion
}

func NuevoGrafo() *Grafo { return &Grafo{nodos:map[string]Nodo{}, relaciones:map[string]Relacion{}} }

func (g *Grafo) AgregarNodo(n Nodo) error {
 if g==nil { return fmt.Errorf("grafo nulo") }
 n.OID=strings.TrimSpace(n.OID)
 if n.OID=="" { return fmt.Errorf("nodo sin oid") }
 g.mu.Lock(); defer g.mu.Unlock()
 if _,ok:=g.nodos[n.OID]; ok { return fmt.Errorf("nodo duplicado: %s",n.OID) }
 g.nodos[n.OID]=n
 return nil
}

func (g *Grafo) AgregarRelacion(r Relacion) error {
 if g==nil { return fmt.Errorf("grafo nulo") }
 if err:=r.Validar(); err!=nil { return err }
 g.mu.Lock(); defer g.mu.Unlock()
 if _,ok:=g.relaciones[r.ID]; ok { return fmt.Errorf("relación duplicada: %s",r.ID) }
 if _,ok:=g.nodos[r.Origen]; !ok { return fmt.Errorf("origen inexistente: %s",r.Origen) }
 if _,ok:=g.nodos[r.Destino]; !ok { return fmt.Errorf("destino inexistente: %s",r.Destino) }
 g.relaciones[r.ID]=r
 return nil
}

func (g *Grafo) Validar() error {
 if g==nil { return fmt.Errorf("grafo nulo") }
 g.mu.RLock(); defer g.mu.RUnlock()
 for id,r:=range g.relaciones {
  if err:=r.Validar(); err!=nil { return err }
  if _,ok:=g.nodos[r.Origen]; !ok { return fmt.Errorf("relación %s: origen inexistente",id) }
  if _,ok:=g.nodos[r.Destino]; !ok { return fmt.Errorf("relación %s: destino inexistente",id) }
 }
 return nil
}

func (g *Grafo) RelacionesDesde(oid string) []Relacion {
 if g==nil { return nil }
 g.mu.RLock(); defer g.mu.RUnlock()
 out:=[]Relacion{}
 for _,r:=range g.relaciones { if r.Origen==oid { out=append(out,r) } }
 sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID})
 return out
}

func (g *Grafo) RelacionesEntre(a,b string) []Relacion {
 if g==nil { return nil }
 g.mu.RLock(); defer g.mu.RUnlock()
 out:=[]Relacion{}
 for _,r:=range g.relaciones { if (r.Origen==a&&r.Destino==b)||(r.Origen==b&&r.Destino==a) { out=append(out,r) } }
 sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID})
 return out
}

// Perfil is descriptive only. It intentionally does not assign a scalar level.
type Perfil struct { Nodos int; Relaciones int; PorTipo map[string]int; PorEvidencia map[Evidencia]int }

func (g *Grafo) Perfil() Perfil {
 p:=Perfil{PorTipo:map[string]int{},PorEvidencia:map[Evidencia]int{}}
 if g==nil { return p }
 g.mu.RLock(); defer g.mu.RUnlock()
 p.Nodos=len(g.nodos); p.Relaciones=len(g.relaciones)
 for _,r:=range g.relaciones {
  p.PorTipo[r.Tipo]++
  e:=r.Evidencia; if e=="" { e=EvidenciaDesconocida }; p.PorEvidencia[e]++
 }
 return p
}

func (g *Grafo) Copia() ([]Nodo,[]Relacion) {
 if g==nil { return nil,nil }
 g.mu.RLock(); defer g.mu.RUnlock()
 ns:=make([]Nodo,0,len(g.nodos)); rs:=make([]Relacion,0,len(g.relaciones))
 for _,n:=range g.nodos { ns=append(ns,n) }; for _,r:=range g.relaciones { rs=append(rs,r) }
 sort.Slice(ns,func(i,j int)bool{return ns[i].OID<ns[j].OID})
 sort.Slice(rs,func(i,j int)bool{return rs[i].ID<rs[j].ID})
 return ns,rs
}
