package declarativo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func snapshotDecl(name string) OrgDecl {
	return OrgDecl{
		OID:"alset:org:snapshot:1", Name:name, Version:"1", Kind:"ORGES",
		Capabilities:[]string{"lectura"},
		InitialState:map[string]interface{}{"contador":1},
	}
}

func TestRuntimeSnapshot_RoundTripVersionado(t *testing.T) {
	dir:=t.TempDir()
	decl:=snapshotDecl("snapshot")
	rt,err:=NewRuntime(dir); if err!=nil {t.Fatal(err)}
	if err:=rt.LoadAssembly([]OrgDecl{decl});err!=nil{t.Fatal(err)}
	rt.SetState("contador",42)
	rt.State.RegisterDeps("resultado",[]string{"contador"})
	rt.State.Dirty["resultado"]=true
	rt.State.EvalRule(decl.OID,Rule{ID:"r",When:[]Condition{{Path:"contador",Op:"gte",Value:1}},ThenPath:"resultado",ThenValue:true})
	if err:=rt.RecordEvaluation("mind.fractal.v1",decl.OID,"r",1,[]byte(`{"estado":1,"nivel":0}`));err!=nil{t.Fatal(err)}
	if err:=rt.Persist();err!=nil{t.Fatal(err)}

	encoded,err:=os.ReadFile(filepath.Join(dir,"runtime_state.json"));if err!=nil{t.Fatal(err)}
	var envelope snapshotEnvelope
	if err:=json.Unmarshal(encoded,&envelope);err!=nil{t.Fatal(err)}
	if envelope.Version!=VersionSnapshot||envelope.Algorithm!="sha256"||envelope.Checksum==""{t.Fatalf("envoltura inválida: %+v",envelope)}

	restored,err:=NewRuntime(dir);if err!=nil{t.Fatal(err)}
	if err:=restored.LoadAssembly([]OrgDecl{decl});err!=nil{t.Fatal(err)}
	if err:=restored.Restore();err!=nil{t.Fatal(err)}
	if value,ok:=restored.GetState("contador");!ok||value!=float64(42){t.Fatalf("átomo restaurado=%v existe=%v",value,ok)}
	dirty:=restored.State.DirtyAtoms()
	if len(dirty)!=1||dirty[0]!="resultado"{t.Fatalf("dirty restaurado=%v",dirty)}
	if len(restored.State.HistoryCopy())!=1{t.Fatalf("historial restaurado=%d",len(restored.State.HistoryCopy()))}
	if len(restored.EvaluationsCopy())!=1{t.Fatalf("evaluaciones restauradas=%d",len(restored.EvaluationsCopy()))}
	if restored.HasCap(decl.OID,"lectura")!=true{t.Fatal("la capacidad declarada no debe perderse al restaurar")}
	if restored.AuditLen()==0{t.Fatal("se esperaba auditoría restaurada")}
}

func TestRuntimeSnapshot_RechazaChecksumCorruptoSinMutar(t *testing.T) {
	dir:=t.TempDir(); decl:=snapshotDecl("checksum")
	rt,_:=NewRuntime(dir);if err:=rt.LoadAssembly([]OrgDecl{decl});err!=nil{t.Fatal(err)}
	rt.SetState("dato","original");if err:=rt.Persist();err!=nil{t.Fatal(err)}
	path:=filepath.Join(dir,"runtime_state.json")
	raw,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)}
	corrupt:=strings.Replace(string(raw),"original","alterado",1)
	if corrupt==string(raw){t.Fatal("no se encontró el valor para corromper")}
	if err:=os.WriteFile(path,[]byte(corrupt),0600);err!=nil{t.Fatal(err)}
	target,_:=NewRuntime(dir);if err:=target.LoadAssembly([]OrgDecl{decl});err!=nil{t.Fatal(err)}
	target.SetState("proteccion","intacto")
	if err:=target.Restore();err==nil||!strings.Contains(err.Error(),"checksum"){t.Fatalf("se esperaba error de checksum, recibido %v",err)}
	if value,_:=target.GetState("proteccion");value!="intacto"{t.Fatalf("restore parcial detectado: %v",value)}
	if _,ok:=target.GetState("dato");ok{t.Fatal("se aplicaron átomos antes de validar checksum")}
}

func TestRuntimeSnapshot_RechazaRootCIDDistintoSinMutar(t *testing.T) {
	dir:=t.TempDir(); first:=snapshotDecl("primero")
	source,_:=NewRuntime(dir);if err:=source.LoadAssembly([]OrgDecl{first});err!=nil{t.Fatal(err)}
	source.SetState("secreto","no-transferir");if err:=source.Persist();err!=nil{t.Fatal(err)}
	second:=snapshotDecl("segundo")
	target,_:=NewRuntime(dir);if err:=target.LoadAssembly([]OrgDecl{second});err!=nil{t.Fatal(err)}
	target.SetState("local","conservar")
	if err:=target.Restore();err==nil||!strings.Contains(err.Error(),"RootCID"){t.Fatalf("se esperaba rechazo por RootCID, recibido %v",err)}
	if value,_:=target.GetState("local");value!="conservar"{t.Fatalf("restore parcial: %v",value)}
	if _,ok:=target.GetState("secreto");ok{t.Fatal("se restauraron datos de otra definición")}
}

func TestRuntimeSnapshot_LegacyValidoSeMigra(t *testing.T) {
	dir:=t.TempDir();decl:=snapshotDecl("legado")
	rt,_:=NewRuntime(dir);if err:=rt.LoadAssembly([]OrgDecl{decl});err!=nil{t.Fatal(err)}
	legacy:=map[string]interface{}{
		"atoms":map[string]interface{}{"dato":"migrado"},
		"root_cids":rt.RootCIDs,
		"bindings":[]Binding{},
		"audit":[]AuditEntry{},
		"history":[]InferenceRecord{},
		"caps":rt.Caps,
	}
	raw,err:=json.Marshal(legacy);if err!=nil{t.Fatal(err)}
	if err:=os.WriteFile(filepath.Join(dir,"runtime_state.json"),raw,0600);err!=nil{t.Fatal(err)}
	if err:=rt.Restore();err!=nil{t.Fatal(err)}
	if value,_:=rt.GetState("dato");value!="migrado"{t.Fatalf("dato legado=%v",value)}
	if err:=rt.Persist();err!=nil{t.Fatal(err)}
	encoded,_:=os.ReadFile(filepath.Join(dir,"runtime_state.json"))
	var envelope snapshotEnvelope
	if err:=json.Unmarshal(encoded,&envelope);err!=nil{t.Fatal(err)}
	if envelope.Version!=VersionSnapshot{t.Fatalf("no se migró a v%d: %+v",VersionSnapshot,envelope)}
}

func TestRuntimeSnapshot_CapacidadesNoSePuedenElevar(t *testing.T) {
	dir:=t.TempDir();decl:=snapshotDecl("capacidades")
	rt,_:=NewRuntime(dir);if err:=rt.LoadAssembly([]OrgDecl{decl});err!=nil{t.Fatal(err)}
	if err:=rt.Persist();err!=nil{t.Fatal(err)}
	path:=filepath.Join(dir,"runtime_state.json")
	raw,_:=os.ReadFile(path)
	var env snapshotEnvelope;_ = json.Unmarshal(raw,&env)
	var payload map[string]interface{};_ = json.Unmarshal(env.Payload,&payload)
	payload["caps"]=map[string]interface{}{decl.OID:map[string]interface{}{"admin":true}}
	canonical,_:=json.Marshal(payload);sum:=sha256Hex(canonical)
	env.Payload=canonical;env.Checksum=sum
	tampered,_:=json.Marshal(env);_ = os.WriteFile(path,tampered,0600)
	target,_:=NewRuntime(dir);if err:=target.LoadAssembly([]OrgDecl{decl});err!=nil{t.Fatal(err)}
	if err:=target.Restore();err==nil||!strings.Contains(err.Error(),"capacidades"){t.Fatalf("se esperaba rechazo de capacidades manipuladas: %v",err)}
	if target.HasCap(decl.OID,"admin"){t.Fatal("snapshot concedió capacidad")}
}

func sha256Hex(data []byte) string {
	sum:=sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
