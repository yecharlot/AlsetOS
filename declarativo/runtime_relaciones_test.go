package declarativo

import (
	"testing"

	"github.com/yecharlot/AlsetOS/relacion"
)

func TestRuntime_RelationalContextReadOnly(t *testing.T) {
	rt, err := NewRuntime(t.TempDir())
	if err != nil { t.Fatal(err) }
	a := OrgDecl{OID:"alset:org:a:1", Name:"A", Version:"1", Kind:"ORGES",
		Relations: []relacion.Relacion{{ID:"a-b", Origen:"alset:org:a:1", Destino:"alset:org:b:1", Tipo:"colabora", Evidencia:relacion.EvidenciaObservada}}}
	b := OrgDecl{OID:"alset:org:b:1", Name:"B", Version:"1", Kind:"ORGES"}
	if err := rt.LoadAssembly([]OrgDecl{a,b}); err != nil { t.Fatal(err) }
	before := rt.AuditLen()
	ctx, err := rt.RelationalContext(a.OID)
	if err != nil { t.Fatal(err) }
	if len(ctx.Salientes)!=1 || ctx.Salientes[0].Destino!=b.OID { t.Fatalf("contexto A: %+v",ctx) }
	ctxB, err := rt.RelationalContext(b.OID)
	if err != nil { t.Fatal(err) }
	if len(ctxB.Entrantes)!=1 || ctxB.Entrantes[0].Origen!=a.OID { t.Fatalf("contexto B: %+v",ctxB) }
	if rt.AuditLen()!=before { t.Fatal("consultar contexto no debe mutar auditoría ni runtime") }
	if rt.HasCap(a.OID,"operar-b") { t.Fatal("una relación no debe conceder capacidades") }
}

func TestRuntime_RelatedOrganismsDeterministic(t *testing.T) {
	rt, err := NewRuntime("")
	if err != nil { t.Fatal(err) }
	a:=OrgDecl{OID:"a",Name:"A",Version:"1",Kind:"ORGES",Relations:[]relacion.Relacion{
		{ID:"r2",Origen:"a",Destino:"c",Tipo:"x"},
	}}
	b:=OrgDecl{OID:"b",Name:"B",Version:"1",Kind:"ORGES",Relations:[]relacion.Relacion{
		{ID:"r1",Origen:"b",Destino:"a",Tipo:"x"},
	}}
	c:=OrgDecl{OID:"c",Name:"C",Version:"1",Kind:"ORGES"}
	if err:=rt.LoadAssembly([]OrgDecl{a,b,c});err!=nil{t.Fatal(err)}
	got,err:=rt.RelatedOrganisms("a");if err!=nil{t.Fatal(err)}
	if len(got)!=2||got[0]!="b"||got[1]!="c"{t.Fatalf("vecinos=%v",got)}
}

func TestRuntime_RelationalContextLegacyAssembly(t *testing.T) {
	rt,err:=NewRuntime("");if err!=nil{t.Fatal(err)}
	d:=OrgDecl{OID:"legacy",Name:"legacy",Version:"1",Kind:"ORGES"}
	if err:=rt.LoadAssembly([]OrgDecl{d});err!=nil{t.Fatal(err)}
	ctx,err:=rt.RelationalContext("legacy");if err!=nil{t.Fatal(err)}
	if len(ctx.Salientes)!=0||len(ctx.Entrantes)!=0{t.Fatalf("contexto=%+v",ctx)}
	profile,err:=rt.RelationalProfile();if err!=nil{t.Fatal(err)}
	if profile.Nodos!=1||profile.Relaciones!=0{t.Fatalf("perfil=%+v",profile)}
}

func TestRuntime_RelationalContextUnknownOrganism(t *testing.T) {
	rt,err:=NewRuntime("");if err!=nil{t.Fatal(err)}
	if err:=rt.LoadAssembly([]OrgDecl{{OID:"a",Name:"A",Version:"1",Kind:"ORGES"}});err!=nil{t.Fatal(err)}
	if _,err:=rt.RelationalContext("absent");err==nil{t.Fatal("se esperaba error para organismo desconocido")}
}
