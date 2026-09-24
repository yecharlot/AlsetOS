package kernel

import(
 "context"
 "time"
 "github.com/yecharlot/AlsetOS/autonomia"
 "github.com/yecharlot/AlsetOS/identidad"
 "github.com/yecharlot/AlsetOS/p2p"
)

type Config struct{
 Escuchar string
 Identidad string
 Organismos string
 DHT string
 Colocacion string
}
type Kernel struct{Nodo *p2p.Nodo;Autonomia *autonomia.Servicio}
func Arrancar(c Config)(*Kernel,error){
 id,err:=cargarIdentidad(c.Identidad);if err!=nil{return nil,err}
 n,err:=p2p.NuevoConEstado(id,c.Escuchar,c.Organismos);if err!=nil{return nil,err}
 if err:=n.ActivarPersistenciaDHT(c.DHT);err!=nil{_ = n.Cerrar();return nil,err}
 a,err:=autonomia.Nuevo(n,c.Colocacion);if err!=nil{_ = n.Cerrar();return nil,err}
 return &Kernel{Nodo:n,Autonomia:a},nil
}
func cargarIdentidad(ruta string)(*identidad.Identidad,error){
 id,err:=identidad.Cargar(ruta)
 if err==nil{return id,nil}
 id,err=identidad.Nueva();if err!=nil{return nil,err}
 if err:=id.Guardar(ruta);err!=nil{return nil,err};return id,nil
}
func(k *Kernel) Ejecutar(ctx context.Context){
 go k.Autonomia.Ejecutar(ctx)
 <-ctx.Done()
}
func(k *Kernel) Cerrar()error{
 if k==nil||k.Nodo==nil{return nil}
 return k.Nodo.Cerrar()
}
func Esperar(ctx context.Context,d time.Duration){if d<=0{<-ctx.Done();return};t:=time.NewTimer(d);defer t.Stop();select{case<-ctx.Done():case<-t.C:}}
