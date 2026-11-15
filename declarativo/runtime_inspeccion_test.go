package declarativo

import "testing"

func TestEvaluateRuleReadOnly_NoMutaRuntime(t *testing.T) {
	rule:=Rule{ID:"lectura",When:[]Condition{{Path:"dato",Op:"eq",Value:"si"}},ThenPath:"salida",ThenValue:true}
	rt,err:=NewRuntime("")
	if err!=nil{t.Fatal(err)}
	if err:=rt.LoadAssembly([]OrgDecl{{OID:"a",Name:"A",Version:"1",Kind:"ORGES",Rules:[]Rule{rule}}});err!=nil{t.Fatal(err)}
	rt.SetState("dato","si")
	beforeAudit:=rt.AuditLen()
	beforeHistory:=len(rt.State.HistoryCopy())
	record,_,err:=rt.EvaluateRuleReadOnly("a","lectura")
	if err!=nil{t.Fatal(err)}
	if record.Zyrion!="si"{t.Fatalf("estado=%s",record.Zyrion)}
	if _,ok:=rt.GetState("salida");ok{t.Fatal("la inspección no debe escribir ThenPath")}
	if rt.AuditLen()!=beforeAudit||len(rt.State.HistoryCopy())!=beforeHistory{t.Fatal("la inspección no debe mutar auditoría ni historial")}
	if _,ok:=rt.State.Deps["salida"];ok{t.Fatal("la inspección no debe registrar dependencias")}
}

func TestRulesFor_ReturnsCopy(t *testing.T) {
	rule:=Rule{ID:"r",When:[]Condition{{Path:"x",Op:"exists"}},ThenPath:"salida"}
	rt,err:=NewRuntime("")
	if err!=nil{t.Fatal(err)}
	if err:=rt.LoadAssembly([]OrgDecl{{OID:"a",Name:"A",Version:"1",Kind:"ORGES",Rules:[]Rule{rule}}});err!=nil{t.Fatal(err)}
	rules,err:=rt.RulesFor("a")
	if err!=nil{t.Fatal(err)}
	rules[0].When[0].Path="mutado"
	again,err:=rt.RulesFor("a")
	if err!=nil{t.Fatal(err)}
	if again[0].When[0].Path!="x"{t.Fatal("RulesFor expuso la definición interna")}
}
