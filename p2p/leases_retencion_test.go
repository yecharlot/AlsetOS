package p2p

import (
	"context"
	"path/filepath"
	"strings"
	"encoding/json"
	"encoding/base64"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestLeaseRetencionDeniegaPorDefectoYAceptaCompromiso(t *testing.T) {
	idA,err:=identidad.Nueva();if err!=nil{t.Fatal(err)}
	idB,err:=identidad.Nueva();if err!=nil{t.Fatal(err)}
	a,err:=Nuevo(idA,"/ip4/127.0.0.1/tcp/0");if err!=nil{t.Fatal(err)};defer a.Cerrar()
	b,err:=Nuevo(idB,"/ip4/127.0.0.1/tcp/0");if err!=nil{t.Fatal(err)};defer b.Cerrar()
	ctx,cancel:=context.WithTimeout(context.Background(),15*time.Second);defer cancel()
	if err:=a.Host.Connect(ctx,peer.AddrInfo{ID:b.ID(),Addrs:b.Host.Network().ListenAddresses()});err!=nil{t.Fatal(err)}
	reconectar:=func(){if err:=a.Host.Connect(ctx,peer.AddrInfo{ID:b.ID(),Addrs:b.Host.Addrs()});err!=nil{t.Fatal(err)}}
	data:=[]byte("contenido que debe quedar retenido")
	root:=rootcid.CrearContenido(data)
	d,err:=CrearDescriptorVersion(idA,"organismo:lease",1,root,"");if err!=nil{t.Fatal(err)}
	if err:=a.almacen.Guardar(root,data);err!=nil{t.Fatal(err)}
	if _,err:=a.versiones.Registrar(d);err!=nil{t.Fatal(err)}
	if _,err:=a.ReplicarDescriptor(ctx,b.ID(),d,data);err!=nil{t.Fatal(err)}
	reconectar();if _,err:=a.SolicitarRetencionRemota(ctx,b.ID(),d,time.Hour);err==nil||!strings.Contains(err.Error(),"no autorizado"){t.Fatalf("rechazo por defecto incorrecto: %v",err)}
	if err:=b.ConfigurarCapacidadRetencion(1);err!=nil{t.Fatal(err)}
	if err:=b.AutorizarRetencionRemota(a.ID(),1<<20);err!=nil{t.Fatal(err)}
	reconectar();if _,err:=a.SolicitarRetencionRemota(ctx,b.ID(),d,time.Hour);err==nil||!strings.Contains(err.Error(),"capacidad"){t.Fatalf("rechazo por capacidad incorrecto: %v",err)}
	if err:=b.ConfigurarCapacidadRetencion(1<<20);err!=nil{t.Fatal(err)}
	if err:=b.AutorizarRetencionRemota(a.ID(),1);err!=nil{t.Fatal(err)}
	reconectar();if _,err:=a.SolicitarRetencionRemota(ctx,b.ID(),d,time.Hour);err==nil||!strings.Contains(err.Error(),"cuota"){t.Fatalf("rechazo por cuota incorrecto: %v",err)}
	if err:=b.AutorizarRetencionRemota(a.ID(),1<<20);err!=nil{t.Fatal(err)}
	reconectar();lease,err:=a.SolicitarRetencionRemota(ctx,b.ID(),d,time.Hour);if err!=nil{t.Fatal(err)}
	if lease.Bytes!=int64(len(data))||lease.Proveedor!=b.ID().String()||!verificarFirmaLease(lease){t.Fatalf("lease devuelto no verificable: %+v",lease)}
	if len(b.LeasesRetencion())!=1{t.Fatalf("lease no persistido en proveedor: %+v",b.LeasesRetencion())}
	reconectar();repetido,err:=a.SolicitarRetencionRemota(ctx,b.ID(),d,time.Hour);if err!=nil{t.Fatalf("repetir solicitud idempotente: %v",err)}
	if repetido.ID!=lease.ID||len(b.LeasesRetencion())!=1{t.Fatalf("la solicitud repetida duplicó el compromiso: primero=%s segundo=%s leases=%+v",lease.ID,repetido.ID,b.LeasesRetencion())}
	reconectar();renovado,err:=a.RenovarLeaseRemoto(ctx,lease,2*time.Hour);if err!=nil{t.Fatal(err)}
	if !renovado.Expira.After(lease.Expira){t.Fatalf("la renovación no extendió el compromiso: anterior=%s nuevo=%s",lease.Expira,renovado.Expira)}
	reconectar();if err:=a.LiberarLeaseRemoto(ctx,renovado);err!=nil{t.Fatalf("liberar lease: %v",err)}
	if len(b.LeasesRetencion())!=0{t.Fatalf("la liberación no retiró el compromiso: %+v",b.LeasesRetencion())}
	alterado:=renovado;alterado.RootCID=rootcid.CrearContenido([]byte("otro contenido"))
	if verificarFirmaLease(alterado){t.Fatal("se aceptó un lease alterado")}
}

func TestGestorLeasesCuotaCapacidadPersistenciaYExpiracion(t *testing.T) {
	id,err:=identidad.Nueva();if err!=nil{t.Fatal(err)}
	data:=[]byte("bloque con cuota")
	root:=rootcid.CrearContenido(data)
	d,err:=CrearDescriptorVersion(id,"organismo:cuota",1,root,"");if err!=nil{t.Fatal(err)}
	path:=filepath.Join(t.TempDir(),"leases.json")
	g,err:=NuevoGestorLeases(path);if err!=nil{t.Fatal(err)}
	now:=time.Now().UTC()
	l:=LeaseRetencion{Schema:EsquemaLeaseRetencion,ID:"lease-test",RootCID:root,OrganismoID:d.OrganismoID,Solicitante:d.AutorPeerID,Proveedor:d.AutorPeerID,Bytes:int64(len(data)),Emitido:now,Expira:now.Add(time.Minute),ClavePublica:d.ClavePublica}
	payload,err:=jsonMarshalLease(l);if err!=nil{t.Fatal(err)}
	l.Firma=signForTest(id,payload)
	g.leases[l.ID]=l
	if err:=g.persistirLocked();err!=nil{t.Fatal(err)}
	reloaded,err:=NuevoGestorLeases(path);if err!=nil{t.Fatal(err)}
	if len(reloaded.Activos(now))!=1{t.Fatal("el lease no sobrevivió al reinicio")}
	if err:=reloaded.ConfigurarCapacidad(l.Bytes-1);err==nil{t.Fatal("se permitió reducir capacidad por debajo del compromiso activo")}
	if err:=reloaded.Expirar(now.Add(2*time.Minute));err!=nil{t.Fatal(err)}
	if len(reloaded.Activos(now.Add(2*time.Minute)))!=0{t.Fatal("el lease vencido sigue activo")}
	final,err:=NuevoGestorLeases(path);if err!=nil{t.Fatal(err)}
	if len(final.Activos(now.Add(2*time.Minute)))!=0{t.Fatal("la expiración no persistió")}
}

func jsonMarshalLease(l LeaseRetencion) ([]byte,error) { return json.Marshal(l.firmable()) }
func signForTest(id *identidad.Identidad,b []byte) string { return base64.StdEncoding.EncodeToString(id.Firmar(b)) }
