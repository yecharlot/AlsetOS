package declarativo

import (
	"strings"
	"testing"

	"github.com/yecharlot/AlsetOS/relacion"
)

func TestResolve_RelacionesEntreOrganismos(t *testing.T) {
	origen := OrgDecl{OID: "alset:org:origen:1", Name: "origen", Version: "1", Kind: "ORGES",
		Relations: []relacion.Relacion{{ID: "r-1", Origen: "alset:org:origen:1", Destino: "alset:org:destino:1", Tipo: "colabora", Evidencia: relacion.EvidenciaObservada}}}
	destino := OrgDecl{OID: "alset:org:destino:1", Name: "destino", Version: "1", Kind: "ORGES"}
	asm, err := Resolve([]OrgDecl{origen, destino})
	if err != nil { t.Fatal(err) }
	got := asm.Relations.RelacionesDesde(origen.OID)
	if len(got) != 1 || got[0].Destino != destino.OID { t.Fatalf("relaciones=%+v", got) }
}

func TestResolve_RechazaDestinoDesconocido(t *testing.T) {
	d := OrgDecl{OID: "alset:org:a:1", Name: "a", Version: "1", Kind: "ORGES",
		Relations: []relacion.Relacion{{ID: "r-1", Origen: "alset:org:a:1", Destino: "alset:org:no-existe:1", Tipo: "depende"}}}
	_, err := Resolve([]OrgDecl{d})
	if err == nil || !strings.Contains(err.Error(), "destino inexistente") { t.Fatalf("error=%v", err) }
}

func TestResolve_RechazaIDDeRelacionDuplicado(t *testing.T) {
	a := OrgDecl{OID: "alset:org:a:1", Name: "a", Version: "1", Kind: "ORGES",
		Relations: []relacion.Relacion{{ID: "misma", Origen: "alset:org:a:1", Destino: "alset:org:b:1", Tipo: "x"}}}
	b := OrgDecl{OID: "alset:org:b:1", Name: "b", Version: "1", Kind: "ORGES",
		Relations: []relacion.Relacion{{ID: "misma", Origen: "alset:org:b:1", Destino: "alset:org:a:1", Tipo: "y"}}}
	_, err := Resolve([]OrgDecl{a, b})
	if err == nil || !strings.Contains(err.Error(), "relación duplicada") { t.Fatalf("error=%v", err) }
}

func TestDefinitionCID_RelacionesOpcionalesYDeterministas(t *testing.T) {
	d := OrgDecl{OID: "alset:org:a:1", Name: "a", Version: "1", Kind: "ORGES"}
	cid1, err := d.DefinitionCID(); if err != nil { t.Fatal(err) }
	d.Relations = []relacion.Relacion{}
	cid2, err := d.DefinitionCID(); if err != nil { t.Fatal(err) }
	if cid1 != cid2 { t.Fatalf("un slice vacío alteró el CID: %s != %s", cid1, cid2) }
	d.Relations = []relacion.Relacion{{ID:"r", Origen:d.OID, Destino:"alset:org:b:1", Tipo:"x"}}
	cid3, err := d.DefinitionCID(); if err != nil { t.Fatal(err) }
	if cid3 == cid1 { t.Fatal("la relación declarada debe formar parte de la identidad de definición") }
}

func TestParsearAOS_RelacionesDeclarativas(t *testing.T) {
	src := `(flow prueba
	  (orge origen
	    (relacion r-stock (destino destino) (tipo dependencia) (direccion dirigida) (evidencia documentada)))
	  (orge destino))`
	flow, err := ParsearAOS(src); if err != nil { t.Fatal(err) }
	if len(flow.Manifiesto.Organismos[0].Relations) != 1 { t.Fatalf("declaraciones=%+v", flow.Manifiesto.Organismos[0].Relations) }
	asm, err := Resolve(flow.Manifiesto.Organismos); if err != nil { t.Fatal(err) }
	if len(asm.Relations.RelacionesDesde("alset:org:origen:1")) != 1 { t.Fatal("la relación no llegó al ensamblaje") }
}

func TestParsearAOS_RelacionConDestinoDesconocido(t *testing.T) {
	src := `(flow prueba (orge origen (relacion r (destino ausente) (tipo dependencia))))`
	flow, err := ParsearAOS(src); if err != nil { t.Fatal(err) }
	_, err = Resolve(flow.Manifiesto.Organismos)
	if err == nil { t.Fatal("se esperaba rechazar la referencia no resuelta") }
}
