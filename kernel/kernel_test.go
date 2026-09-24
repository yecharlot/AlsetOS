package kernel

import(
 "path/filepath"
 "testing"
)

func TestArrancarCerrar(t *testing.T){
 dir:=t.TempDir()
 k,err:=Arrancar(Config{
  Escuchar:"/ip4/127.0.0.1/tcp/0",
  Identidad:filepath.Join(dir,"identidad.json"),
  Organismos:filepath.Join(dir,"organismos.json"),
  DHT:filepath.Join(dir,"dht"),
  Colocacion:filepath.Join(dir,"colocacion.json"),
 })
 if err!=nil{t.Fatal(err)}
 if k.Nodo.ID().String()==""{t.Fatal("node id vacío")}
 if err:=k.Cerrar();err!=nil{t.Fatal(err)}
}
