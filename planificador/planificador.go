package planificador
import("context";"sync";"time")
type Tarea struct{Name string;Intervalo time.Duration;Accion func(context.Context)error}
type Planificador struct{mu sync.Mutex;tareas []Tarea}
func Nuevo()*Planificador{return &Planificador{}}
func(p *Planificador)Registrar(t Tarea){p.mu.Lock();defer p.mu.Unlock();p.tareas=append(p.tareas,t)}
func(p *Planificador)Ejecutar(ctx context.Context){p.mu.Lock();ts:=append([]Tarea(nil),p.tareas...);p.mu.Unlock();var wg sync.WaitGroup;for _,t:=range ts{if t.Intervalo<=0||t.Accion==nil{continue};t:=t;wg.Add(1);go func(){defer wg.Done();tick:=time.NewTicker(t.Intervalo);defer tick.Stop();for{select{case<-ctx.Done():return;case<-tick.C:_=t.Accion(ctx)}}}()};<-ctx.Done();wg.Wait()}
