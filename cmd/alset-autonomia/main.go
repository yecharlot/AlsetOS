package main

import(
 "context"
 "flag"
 "fmt"
 "os"
 "time"
 "github.com/libp2p/go-libp2p/core/peer"
 "github.com/multiformats/go-multiaddr"
 "github.com/yecharlot/AlsetOS/autonomia"
 "github.com/yecharlot/AlsetOS/identidad"
 "github.com/yecharlot/AlsetOS/manifiesto"
 "github.com/yecharlot/AlsetOS/p2p"
 "github.com/yecharlot/AlsetOS/rootcid"
)
func cargarIdentidad(ruta string)(*identidad.Identidad,error){if _,err:=os.Stat(ruta);err==nil{return identidad.Cargar(ruta)}else if !os.IsNotExist(err){return nil,err};n,err:=identidad.Nueva();if err!=nil{return nil,err};if err:=n.Guardar(ruta);err!=nil{return nil,err};return n,nil}
func conectar(ctx context.Context,nodo *p2p.Nodo,direccion string)error{if direccion==""{return nil};m,err:=multiaddr.NewMultiaddr(direccion);if err!=nil{return err};info,err:=peer.AddrInfoFromP2pAddr(m);if err!=nil{return err};return nodo.Host.Connect(ctx,*info)}
func main(){
 escuchar:=flag.String("escuchar","/ip4/127.0.0.1/tcp/0","dirección libp2p")
 peerRemoto:=flag.String("peer","","multiaddr /p2p del nodo remoto")
 identidadRuta:=flag.String("identidad","estado/identidad-p2p.json","identidad persistente")
 organismosRuta:=flag.String("organismos","estado/organismos.json","almacén de organismos")
 dhtRuta:=flag.String("dht","estado/dht","datastore persistente de DHT")
 colocacionRuta:=flag.String("colocacion","estado/colocacion.json","registro de placement")
 publicar:=flag.String("publicar","","manifiesto .alset a publicar y replicar")
 replicas:=flag.Int("replicas",2,"cantidad de réplicas deseadas")
 duracion:=flag.Duration("duracion",30*time.Second,"duración del monitor")
 flag.Parse()
 id,err:=cargarIdentidad(*identidadRuta);if err!=nil{fmt.Printf("error identidad: %v
",err);os.Exit(1)}
 nodo,err:=p2p.NuevoConEstado(id,*escuchar,*organismosRuta);if err!=nil{fmt.Printf("error nodo: %v
",err);os.Exit(1)};defer nodo.Cerrar()
 if err:=nodo.ActivarPersistenciaDHT(*dhtRuta);err!=nil{fmt.Printf("error DHT persistente: %v
",err);os.Exit(1)}
 ctx,cancel:=context.WithTimeout(context.Background(),15*time.Second);defer cancel()
 if err:=conectar(ctx,nodo,*peerRemoto);err!=nil{fmt.Printf("error peer: %v
",err);os.Exit(1)}
 if *peerRemoto!=""{if err:=nodo.BootstrapDHT(ctx);err!=nil{fmt.Printf("error DHT: %v
",err);os.Exit(1)}}
 servicio,err:=autonomia.Nuevo(nodo,*colocacionRuta);if err!=nil{fmt.Printf("error autonomía: %v
",err);os.Exit(1)};servicio.FactorReplica=*replicas
 if *publicar!=""{doc,err:=manifiesto.Cargar(*publicar);if err!=nil{fmt.Printf("error manifiesto: %v
",err);os.Exit(1)};contenido,err:=doc.Canonico();if err!=nil{fmt.Printf("error canónico: %v
",err);os.Exit(1)};root:=rootcid.CrearContenido(contenido);c,err:=servicio.Publicar(ctx,root,contenido);if err!=nil{fmt.Printf("error publicar: %v
",err);os.Exit(1)};fmt.Printf("[ALSET-PLACEMENT] rootcid=%s primario=%s replicas=%v
",c.RootCID,c.Primario,c.Replicas)}
 fmt.Printf("[ALSET-AUTONOMIA] node=%s intervalo=%s timeout=%s dht=%s
",nodo.ID(),servicio.Intervalo,servicio.Timeout,*dhtRuta)
 monitorCtx,monitorCancel:=context.WithCancel(context.Background());defer monitorCancel();go servicio.Ejecutar(monitorCtx)
 timer:=time.NewTimer(*duracion);defer timer.Stop();<-timer.C
 for _,c:=range servicio.Registro.Todos(){fmt.Printf("[ALSET-PLACEMENT] rootcid=%s primario=%s replicas=%v
",c.RootCID,c.Primario,c.Replicas)}
}
