package p2p

import (
 "fmt"
 "os"
 "path/filepath"
 "github.com/ipfs/go-ds-pebble"
 kaddht "github.com/libp2p/go-libp2p-kad-dht"
 datastore "github.com/ipfs/go-datastore"
 syncds "github.com/ipfs/go-datastore/sync"
)

func(nodo *Nodo) ActivarPersistenciaDHT(ruta string)error{
 if nodo==nil{return fmt.Errorf("nodo nulo")}
 if ruta==""{return fmt.Errorf("ruta DHT vacía")}
 if err:=os.MkdirAll(filepath.Dir(ruta),0700);err!=nil{return fmt.Errorf("crear directorio DHT: %w",err)}
 if nodo.DHT!=nil{if err:=nodo.DHT.Close();err!=nil{return fmt.Errorf("cerrar DHT temporal: %w",err)}}
 store,err:=pebbleds.NewDatastore(ruta)
 if err!=nil{return fmt.Errorf("abrir datastore DHT: %w",err)}
 nodo.datastore=syncds.MutexWrap(store)
 dht,err:=kaddht.New(nodo.Host,kaddht.Datastore(nodo.datastore),kaddht.Mode(kaddht.ModeServer))
 if err!=nil{_ = store.Close();return fmt.Errorf("recrear DHT persistente: %w",err)}
 nodo.DHT=dht
 return nil
}

func(nodo *Nodo) CerrarPersistencia()error{
 if nodo==nil{return nil}
 if nodo.DHT!=nil{return nodo.DHT.Close()}
 return nil
}

var _ datastore.Batching = (datastore.Batching)(nil)
