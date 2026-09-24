package main

import(
 "context"
 "flag"
 "fmt"
 "os"
 "os/signal"
 "syscall"
 "github.com/yecharlot/AlsetOS/kernel"
)

func main(){
 escuchar:=flag.String("escuchar","/ip4/127.0.0.1/tcp/0","dirección libp2p")
 identidad:=flag.String("identidad","estado/identidad-p2p.json","identidad")
 organismos:=flag.String("organismos","estado/organismos.json","organismos")
 dht:=flag.String("dht","estado/dht","DHT")
 colocacion:=flag.String("colocacion","estado/colocacion.json","placement")
 flag.Parse()
 k,err:=kernel.Arrancar(kernel.Config{Escuchar:*escuchar,Identidad:*identidad,Organismos:*organismos,DHT:*dht,Colocacion:*colocacion})
 if err!=nil{fmt.Printf("error kernel: %v
",err);os.Exit(1)}
 defer k.Cerrar()
 fmt.Printf("[ALSET-KERNEL] node=%s
",k.Nodo.ID())
 fmt.Printf("[ALSET-KERNEL] direcciones=%v
",k.Nodo.Direcciones())
 ctx,cancel:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM);defer cancel()
 k.Ejecutar(ctx)
}
