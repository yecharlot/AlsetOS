package p2p

import (
 "path/filepath"
 "testing"
 "github.com/yecharlot/AlsetOS/identidad"
)

func TestDHTPersistenteReabre(t *testing.T){
 id,err:=identidad.Nueva();if err!=nil{t.Fatal(err)}
 ruta:=filepath.Join(t.TempDir(),"dht")
 n,err:=Nuevo(id,"/ip4/127.0.0.1/tcp/0");if err!=nil{t.Fatal(err)}
 if err:=n.ActivarPersistenciaDHT(ruta);err!=nil{t.Fatal(err)}
 if err:=n.Cerrar();err!=nil{t.Fatal(err)}
 n2,err:=Nuevo(id,"/ip4/127.0.0.1/tcp/0");if err!=nil{t.Fatal(err)}
 if err:=n2.ActivarPersistenciaDHT(ruta);err!=nil{t.Fatal(err)}
 if err:=n2.Cerrar();err!=nil{t.Fatal(err)}
}
