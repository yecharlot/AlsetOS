package mind

import (
	"testing"

	"github.com/yecharlot/AlsetOS/declarativo"
	"github.com/yecharlot/AlsetOS/relacion"
	"github.com/yecharlot/AlsetOS/zyrion"
)

func runtimeFractal(t *testing.T, declarations ...declarativo.OrgDecl) *declarativo.Runtime {
	t.Helper()
	rt, err := declarativo.NewRuntime("")
	if err != nil { t.Fatal(err) }
	if err := rt.LoadAssembly(declarations); err != nil { t.Fatal(err) }
	return rt
}

func reglaFractal() declarativo.Rule {
	return declarativo.Rule{ID:"misma-regla", When:[]declarativo.Condition{{Path:"dato", Op:"eq", Value:"listo"}}, ThenPath:"resultado", ThenValue:true}
}

func TestEvaluarFractal_IndeterminadoResueltoEnNivelSiguiente(t *testing.T) {
	rule := reglaFractal()
	a := declarativo.OrgDecl{OID:"a",Name:"A",Version:"1",Kind:"ORGES",Rules:[]declarativo.Rule{rule},
		Relations:[]relacion.Relacion{{ID:"a-b",Origen:"a",Destino:"b",Tipo:TipoRelacionFractal}}}
	b := declarativo.OrgDecl{OID:"b",Name:"B",Version:"1",Kind:"ORGES",Rules:[]declarativo.Rule{rule},
		InitialState:map[string]interface{}{"b.dato":"listo"}}
	rt := runtimeFractal(t,a,b)
	beforeAudit, beforeHistory := rt.AuditLen(), len(rt.State.HistoryCopy())
	got, err := (Mente{}).EvaluarFractal(rt,"a",rule.ID,4)
	if err != nil { t.Fatal(err) }
	if got.Estado != zyrion.Si || got.Raiz.Estado != zyrion.Si { t.Fatalf("resultado: %+v",got) }
	if len(got.Raiz.Hijos)!=1 || got.Raiz.Hijos[0].Estado!=zyrion.Si { t.Fatalf("traza: %+v",got.Raiz) }
	if rt.AuditLen()!=beforeAudit || len(rt.State.HistoryCopy())!=beforeHistory { t.Fatal("la evaluación fractal debe ser de solo lectura") }
	if rt.HasCap("a","gene.ejecutar") { t.Fatal("la inferencia no debe conceder capacidades") }
	records := rt.EvaluationsCopy()
	if len(records) != 1 || records[0].Schema != "mind.fractal.v1" || records[0].State != int(zyrion.Si) { t.Fatalf("registro fractal no conservado: %+v", records) }
}

func TestEvaluarFractal_ConflictoPermaneceIndeterminado(t *testing.T) {
	rule:=reglaFractal()
	a:=declarativo.OrgDecl{OID:"a",Name:"A",Version:"1",Kind:"ORGES",Rules:[]declarativo.Rule{rule},
		Relations:[]relacion.Relacion{{ID:"a-b",Origen:"a",Destino:"b",Tipo:TipoRelacionFractal},{ID:"a-c",Origen:"a",Destino:"c",Tipo:TipoRelacionFractal}}}
	b:=declarativo.OrgDecl{OID:"b",Name:"B",Version:"1",Kind:"ORGES",Rules:[]declarativo.Rule{rule},InitialState:map[string]interface{}{"b.dato":"listo"}}
	c:=declarativo.OrgDecl{OID:"c",Name:"C",Version:"1",Kind:"ORGES",Rules:[]declarativo.Rule{rule},InitialState:map[string]interface{}{"c.dato":"no"}}
	rt:=runtimeFractal(t,a,b,c)
	got,err:=(Mente{}).EvaluarFractal(rt,"a",rule.ID,4)
	if err!=nil{t.Fatal(err)}
	if got.Estado!=zyrion.Incierto{t.Fatalf("se esperaba 2 ante conflicto; obtenido %v",got.Estado)}
}

func TestEvaluarFractal_AsociacionNormalNoActivaExpansion(t *testing.T) {
	rule:=reglaFractal()
	a:=declarativo.OrgDecl{OID:"a",Name:"A",Version:"1",Kind:"ORGES",Rules:[]declarativo.Rule{rule},
		Relations:[]relacion.Relacion{{ID:"a-b",Origen:"a",Destino:"b",Tipo:"colabora"}}}
	b:=declarativo.OrgDecl{OID:"b",Name:"B",Version:"1",Kind:"ORGES",Rules:[]declarativo.Rule{rule},InitialState:map[string]interface{}{"b.dato":"listo"}}
	rt:=runtimeFractal(t,a,b)
	got,err:=(Mente{}).EvaluarFractal(rt,"a",rule.ID,4)
	if err!=nil{t.Fatal(err)}
	if got.Estado!=zyrion.Incierto || len(got.Raiz.Hijos)!=0 { t.Fatalf("una relación común no debe activar expansión fractal: %+v",got) }
}

func TestEvaluarFractal_ReglaNoAplicable(t *testing.T) {
	rule:=reglaFractal()
	a:=declarativo.OrgDecl{OID:"a",Name:"A",Version:"1",Kind:"ORGES",Rules:[]declarativo.Rule{rule},
		Relations:[]relacion.Relacion{{ID:"a-b",Origen:"a",Destino:"b",Tipo:TipoRelacionFractal}}}
	b:=declarativo.OrgDecl{OID:"b",Name:"B",Version:"1",Kind:"ORGES"}
	rt:=runtimeFractal(t,a,b)
	got,err:=(Mente{}).EvaluarFractal(rt,"a",rule.ID,1)
	if err!=nil{t.Fatal(err)}
	if got.Estado!=zyrion.Incierto || got.Raiz.Motivo!="estado 2: no existe un nivel fractal relacionado con la misma regla" { t.Fatalf("resultado inesperado: %+v",got) }
}

func TestEvaluarFractal_OrganismoDesconocido(t *testing.T) {
	rt:=runtimeFractal(t,declarativo.OrgDecl{OID:"a",Name:"A",Version:"1",Kind:"ORGES"})
	if _,err:=(Mente{}).EvaluarFractal(rt,"missing","x",2);err==nil{t.Fatal("se esperaba error")}
}
