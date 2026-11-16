package p2p

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestAlmacenContenidoReparaBloqueCorrupto(t *testing.T) {
	dir:=t.TempDir()
	almacen,err:=NuevoAlmacenContenido(dir);if err!=nil{t.Fatal(err)}
	contenido:=[]byte("bloque original y verificable")
	root:=rootcid.CrearContenido(contenido)
	if err:=almacen.Guardar(root,contenido);err!=nil{t.Fatal(err)}
	digest:=root[len("rootcid:"):]
	path:=filepath.Join(dir,digest[:2],digest)
	if err:=os.WriteFile(path,[]byte("datos corruptos"),0600);err!=nil{t.Fatal(err)}
	if err:=almacen.Reparar(root,contenido);err!=nil{t.Fatal(err)}
	recuperado,err:=almacen.Obtener(root);if err!=nil{t.Fatal(err)}
	if string(recuperado)!=string(contenido){t.Fatalf("contenido reparado incorrecto: %q",recuperado)}
}

func TestGestorRetencionPersisteYLibera(t *testing.T) {
	id,err:=identidad.Nueva();if err!=nil{t.Fatal(err)}
	contenido:=[]byte("contenido fijado")
	root:=rootcid.CrearContenido(contenido)
	d,err:=CrearDescriptorVersion(id,"organismo:retencion",1,root,"");if err!=nil{t.Fatal(err)}
	path:=filepath.Join(t.TempDir(),"retenciones.json")
	g,err:=NuevoGestorRetencion(path);if err!=nil{t.Fatal(err)}
	ahora:=time.Now().UTC()
	if err:=g.FijarVersion(d,3,ahora);err!=nil{t.Fatal(err)}
	reabierto,err:=NuevoGestorRetencion(path);if err!=nil{t.Fatal(err)}
	registros:=reabierto.Todas()
	if len(registros)!=1||registros[0].Factor!=3||registros[0].Descriptor.RootCID!=root{t.Fatalf("retención no persistida: %+v",registros)}
	if err:=reabierto.Liberar(root);err!=nil{t.Fatal(err)}
	ultimo,err:=NuevoGestorRetencion(path);if err!=nil{t.Fatal(err)}
	if len(ultimo.Todas())!=0{t.Fatalf("la retención no se liberó: %+v",ultimo.Todas())}
}
