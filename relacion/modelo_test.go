package relacion

import "testing"

func TestGrafoAgregaYConsulta(t *testing.T) {
 g:=NuevoGrafo()
 for _,n:=range []Nodo{{OID:"persona"},{OID:"sueno"},{OID:"actividad"}} { if err:=g.AgregarNodo(n); err!=nil { t.Fatal(err) } }
 if err:=g.AgregarRelacion(Relacion{ID:"r1",Origen:"sueno",Destino:"actividad",Tipo:"asociacion",Evidencia:EvidenciaHipotesis}); err!=nil { t.Fatal(err) }
 if err:=g.Validar(); err!=nil { t.Fatal(err) }
 if len(g.RelacionesDesde("sueno"))!=1 { t.Fatal("consulta de relaciones incorrecta") }
 p:=g.Perfil()
 if p.Nodos!=3 || p.Relaciones!=1 || p.PorTipo["asociacion"]!=1 || p.PorEvidencia[EvidenciaHipotesis]!=1 { t.Fatalf("perfil: %+v",p) }
}

func TestRechazaExtremosInexistentes(t *testing.T) {
 g:=NuevoGrafo(); _=g.AgregarNodo(Nodo{OID:"a"})
 if err:=g.AgregarRelacion(Relacion{ID:"r",Origen:"a",Destino:"b",Tipo:"depende_de"}); err==nil { t.Fatal("debía rechazar destino inexistente") }
}

func TestRechazaDuplicadosYEvidenciaInvalida(t *testing.T) {
 g:=NuevoGrafo(); _=g.AgregarNodo(Nodo{OID:"a"})
 if err:=g.AgregarNodo(Nodo{OID:"a"}); err==nil { t.Fatal("debía rechazar nodo duplicado") }
 if err:=(Relacion{ID:"r",Origen:"a",Destino:"b",Tipo:"x",Evidencia:"inventada"}).Validar(); err==nil { t.Fatal("debía rechazar evidencia inválida") }
}

func TestCopiaOrdenada(t *testing.T) {
 g:=NuevoGrafo(); _=g.AgregarNodo(Nodo{OID:"b"}); _=g.AgregarNodo(Nodo{OID:"a"})
 _=g.AgregarRelacion(Relacion{ID:"z",Origen:"a",Destino:"b",Tipo:"conecta"})
 ns,rs:=g.Copia()
 if ns[0].OID!="a" || rs[0].ID!="z" { t.Fatal("copia no determinista") }
}
