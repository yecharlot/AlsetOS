package motor

import(
 "context"
 "fmt"
 "time"
 "github.com/yecharlot/AlsetOS/agente"
 "github.com/yecharlot/AlsetOS/capacidad"
 "github.com/yecharlot/AlsetOS/eventos"
 "github.com/yecharlot/AlsetOS/gene"
 "github.com/yecharlot/AlsetOS/lispai"
 "github.com/yecharlot/AlsetOS/manifiesto"
 "github.com/yecharlot/AlsetOS/memoria"
 "github.com/yecharlot/AlsetOS/mind"
 "github.com/yecharlot/AlsetOS/organismo"
 "github.com/yecharlot/AlsetOS/pulso"
 "github.com/yecharlot/AlsetOS/recursos"
 "github.com/yecharlot/AlsetOS/rootcid"
 "github.com/yecharlot/AlsetOS/wasm"
)
type Resultado struct{Name string;RootCID string;Estado string;Decision string;GenesEjecutados []string;WASMResultados []uint64;AgentesEjecutados []string;UltimoPulso string;EventoHash string}
type Motor struct{RegistroGenes *gene.Registro;Eventos *eventos.Registro;WASM *wasm.Ejecutor}
func Nuevo()*Motor{registro:=gene.NuevoRegistro();registro.Registrar(gene.Gene{Nombre:"gene-saludo"});registro.Registrar(gene.Gene{Nombre:"gene-sandbox"});return &Motor{RegistroGenes:registro,WASM:wasm.Nuevo()}}
func NuevoConEventos(ruta string)(*Motor,error){m:=Nuevo();r,err:=eventos.Nuevo(ruta);if err!=nil{return nil,err};m.Eventos=r;return m,nil}
func(motor *Motor)auditar(e eventos.Evento){if motor.Eventos!=nil{_,_=motor.Eventos.Emitir(e)}}
func(motor *Motor)EjecutarManifiesto(d manifiesto.Manifiesto)(Resultado,error){
 if err:=d.Validar();err!=nil{return Resultado{},err};estado,err:=d.EstadoZyrion();if err!=nil{return Resultado{},err};contenido,err:=d.Canonico();if err!=nil{return Resultado{},err}
 entidad:=organismo.Nuevo(d.Nombre);entidad.RootCID=rootcid.CrearContenido(contenido);entidad.Estado=organismo.Listo
 motor.auditar(eventos.Evento{Tipo:"organismo.inicio",RootCID:entidad.RootCID,Origen:entidad.Nombre,Contenido:"crear"})
 if d.Memoria!=""{datos,err:=memoria.Cargar(d.Memoria);if err!=nil{entidad.Estado=organismo.Error;motor.auditar(eventos.Evento{Tipo:"organismo.error",RootCID:entidad.RootCID,Contenido:err.Error()});return Resultado{},err};entidad.Memoria=datos}
 for _,c:=range d.Capacidades{entidad.Capacidad[c]=true};for _,r:=range d.Recursos{recursos.Conceder(entidad,r)}
 interprete:=lispai.Interprete{};for _,programa:=range d.LispAI{interprete.Evaluar(programa,entidad)}
 mente:=mind.Mente{};estrategia:=mind.Detener;if d.Zyrion.Estrategia==string(mind.Evaluar){estrategia=mind.Evaluar};decision:=mente.DecidirConEstrategia(estado,entidad,estrategia)
 resultado:=Resultado{Name:entidad.Nombre,RootCID:entidad.RootCID,Estado:estado.Texto(),Decision:decision}
 motor.auditar(eventos.Evento{Tipo:"mind.decision",RootCID:entidad.RootCID,Origen:"mind",Contenido:decision})
 if decision=="evaluar_incierto"{decision=mind.EvaluarIncierto(entidad);resultado.Decision=decision;motor.auditar(eventos.Evento{Tipo:"mind.evaluacion",RootCID:entidad.RootCID,Origen:"mind",Contenido:decision})}
 if decision=="ejecutar_gene"{if err:=capacidad.Requerir(entidad,"gene.ejecutar");err!=nil{entidad.Estado=organismo.Error;return Resultado{},err};entidad.Estado=organismo.Ejecutando;for _,n:=range d.Genes{g,ok:=motor.RegistroGenes.Obtener(n);if !ok{entidad.Estado=organismo.Error;return Resultado{},fmt.Errorf("Gene no registrado: %s",n)};r:=g.Ejecutar(entidad);resultado.GenesEjecutados=append(resultado.GenesEjecutados,r);motor.auditar(eventos.Evento{Tipo:"gene.ejecutado",RootCID:entidad.RootCID,Origen:n,Contenido:r})}}
 if decision=="ejecutar_gene"&&len(d.WASM)>0{if err:=capacidad.Requerir(entidad,"wasm.ejecutar");err!=nil{return Resultado{},err};for _,modulo:=range d.WASM{valor,err:=motor.WASM.Ejecutar(context.Background(),modulo.Ruta,modulo.Funcion);if err!=nil{return Resultado{},err};resultado.WASMResultados=append(resultado.WASMResultados,valor);motor.auditar(eventos.Evento{Tipo:"wasm.ejecutado",RootCID:entidad.RootCID,Origen:modulo.Ruta,Contenido:fmt.Sprint(valor)})}}
 for _,n:=range d.Agentes{if err:=capacidad.Requerir(entidad,"agente.ejecutar");err!=nil{entidad.Estado=organismo.Error;return Resultado{},err};actor:=agente.Agente{Nombre:n,Objetivo:decision};actor.Actuar(entidad,"procesar:"+decision);resultado.AgentesEjecutados=append(resultado.AgentesEjecutados,n)}
 pulso.Emitir(entidad,pulso.Pulso{Tipo:"resultado",Origen:entidad.RootCID,Contenido:decision,Fecha:time.Now()});resultado.UltimoPulso=entidad.Memoria["ultimo_pulso"];entidad.Estado=organismo.Detenido
 if d.Eventos!=""&&motor.Eventos==nil{r,err:=eventos.Nuevo(d.Eventos);if err!=nil{return Resultado{},err};motor.Eventos=r}
 if motor.Eventos!=nil{e,err:=motor.Eventos.Emitir(eventos.Evento{Tipo:"organismo.resultado",RootCID:entidad.RootCID,Origen:entidad.Nombre,Contenido:resultado.UltimoPulso});if err==nil{resultado.EventoHash=e.Hash}}
 if d.Memoria!=""{if err:=memoria.Guardar(d.Memoria,entidad.Memoria);err!=nil{entidad.Estado=organismo.Error;return Resultado{},err}}
 return resultado,nil
}
