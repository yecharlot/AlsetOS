package p2p

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestRepararRetencionesRestauraReplicaRemotaYBloqueLocal(t *testing.T) {
	idA,err:=identidad.Nueva();if err!=nil{t.Fatal(err)}
	idB,err:=identidad.Nueva();if err!=nil{t.Fatal(err)}
	dirA,dirB:=t.TempDir(),t.TempDir()
	nodoA,err:=NuevoConEstado(idA,"/ip4/127.0.0.1/tcp/0",filepath.Join(dirA,"organismos.json"));if err!=nil{t.Fatal(err)}
	defer nodoA.Cerrar()
	nodoB,err:=NuevoConEstado(idB,"/ip4/127.0.0.1/tcp/0",filepath.Join(dirB,"organismos.json"));if err!=nil{t.Fatal(err)}
	defer nodoB.Cerrar()
	ctx,cancelar:=context.WithTimeout(context.Background(),25*time.Second);defer cancelar()
	if err:=nodoA.Host.Connect(ctx,peer.AddrInfo{ID:nodoB.ID(),Addrs:nodoB.Host.Addrs()});err!=nil{t.Fatal(err)}

	contenido:=[]byte("contenido fijado con reparación verificable")
	root:=rootcid.CrearContenido(contenido)
	d,err:=CrearDescriptorVersion(idA,"organismo:retencion-reparacion",1,root,"");if err!=nil{t.Fatal(err)}
	if err:=nodoA.almacen.Guardar(root,contenido);err!=nil{t.Fatal(err)}
	if _,err:=nodoA.versiones.Registrar(d);err!=nil{t.Fatal(err)}
	if err:=nodoA.FijarVersion(d,2);err!=nil{t.Fatal(err)}
	if _,err:=nodoA.ReplicarDescriptor(ctx,nodoB.ID(),d,contenido);err!=nil{t.Fatal(err)}

	digest:=root[len("rootcid:"):]
	pathB:=filepath.Join(nodoB.almacen.dir,digest[:2],digest)
	if err:=os.Remove(pathB);err!=nil{t.Fatal(err)}
	estados,err:=nodoA.RepararRetenciones(ctx)
	if err!=nil{t.Fatalf("reparación de réplica remota: estados=%+v err=%v",estados,err)}
	if ok,err:=nodoB.DisponibleLocal(root);err!=nil||!ok{t.Fatalf("peer B no recuperó réplica: disponible=%v err=%v",ok,err)}

	// Corromper la copia fijada local: el nodo debe recuperar desde el peer
	// y reemplazarla únicamente después de validar el RootCID.
	pathA:=filepath.Join(nodoA.almacen.dir,digest[:2],digest)
	if err:=os.WriteFile(pathA,[]byte("corrupción deliberada"),0600);err!=nil{t.Fatal(err)}
	if _,err:=nodoA.almacen.Obtener(root);err==nil{t.Fatal("el almacén no detectó corrupción local")}
	if _,err:=nodoA.RepararRetenciones(ctx);err!=nil{t.Fatalf("reparación de bloque local: %v",err)}
	recuperado,err:=nodoA.almacen.Obtener(root);if err!=nil{t.Fatal(err)}
	if string(recuperado)!=string(contenido){t.Fatalf("bloque local restaurado incorrectamente: %q",recuperado)}
}

func TestMantenimientoRechazaIntervaloAgresivo(t *testing.T) {
	id,err:=identidad.Nueva();if err!=nil{t.Fatal(err)}
	nodo,err:=Nuevo(id,"/ip4/127.0.0.1/tcp/0");if err!=nil{t.Fatal(err)}
	defer nodo.Cerrar()
	if err:=nodo.IniciarMantenimiento(10*time.Second);err==nil{t.Fatal("aceptó intervalo menor a un minuto")}
	if err:=nodo.IniciarMantenimiento(time.Minute);err!=nil{t.Fatalf("no inició mantenimiento: %v",err)}
	if err:=nodo.IniciarMantenimiento(time.Minute);err==nil{t.Fatal("permitió iniciar dos workers")}
}
