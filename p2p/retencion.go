package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/rootcid"
)

const (
	EsquemaRetencion = "alset.content-retention.v1"
	maxFactorRetencion = 32
)

type RetencionVersion struct {
	Schema string `json:"schema"`
	Descriptor DescriptorVersion `json:"descriptor"`
	Factor int `json:"factor"`
	FijadoEn time.Time `json:"fijado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

type GestorRetencion struct {
	mu sync.RWMutex
	ruta string
	registros map[string]RetencionVersion
}

func NuevoGestorRetencion(ruta string) (*GestorRetencion,error) {
	g:=&GestorRetencion{ruta:ruta,registros:make(map[string]RetencionVersion)}
	if ruta=="" { return g,nil }
	b,err:=os.ReadFile(ruta)
	if os.IsNotExist(err) { return g,nil }
	if err!=nil { return nil,fmt.Errorf("leer política de retención: %w",err) }
	if len(b)==0 { return g,nil }
	var registros map[string]RetencionVersion
	if err:=json.Unmarshal(b,&registros); err!=nil { return nil,fmt.Errorf("decodificar política de retención: %w",err) }
	for root,r:=range registros {
		if r.Schema!=EsquemaRetencion || r.Descriptor.RootCID!=root || r.Factor<1 || r.Factor>maxFactorRetencion {
			return nil,fmt.Errorf("registro de retención inválido para %s",root)
		}
		if err:=VerificarDescriptor(r.Descriptor); err!=nil { return nil,fmt.Errorf("descriptor retenido inválido para %s: %w",root,err) }
		if r.FijadoEn.IsZero() || r.ActualizadoEn.Before(r.FijadoEn) { return nil,fmt.Errorf("fechas de retención inválidas para %s",root) }
	}
	if registros!=nil { g.registros=registros }
	return g,nil
}

// FijarVersion declara una política local explícita. No solicita almacenamiento
// remoto ni interpreta la firma del autor como autorización.
func (g *GestorRetencion) FijarVersion(d DescriptorVersion,factor int,ahora time.Time) error {
	if g==nil { return fmt.Errorf("gestor de retención nulo") }
	if factor<1 || factor>maxFactorRetencion { return fmt.Errorf("factor de retención fuera del rango 1..%d",maxFactorRetencion) }
	if err:=VerificarDescriptor(d); err!=nil { return err }
	g.mu.Lock(); defer g.mu.Unlock()
	anterior,existe:=g.registros[d.RootCID]
	fijado:=ahora.UTC()
	if existe { fijado=anterior.FijadoEn }
	g.registros[d.RootCID]=RetencionVersion{Schema:EsquemaRetencion,Descriptor:d,Factor:factor,FijadoEn:fijado,ActualizadoEn:ahora.UTC()}
	if err:=g.persistirLocked(); err!=nil {
		if existe { g.registros[d.RootCID]=anterior } else { delete(g.registros,d.RootCID) }
		return err
	}
	return nil
}

func (g *GestorRetencion) Liberar(root string) error {
	if g==nil { return fmt.Errorf("gestor de retención nulo") }
	g.mu.Lock(); defer g.mu.Unlock()
	anterior,existe:=g.registros[root]
	if !existe { return nil }
	delete(g.registros,root)
	if err:=g.persistirLocked(); err!=nil { g.registros[root]=anterior; return err }
	return nil
}

func (g *GestorRetencion) Todas() []RetencionVersion {
	if g==nil { return nil }
	g.mu.RLock(); defer g.mu.RUnlock()
	out:=make([]RetencionVersion,0,len(g.registros))
	for _,r:=range g.registros { out=append(out,r) }
	sort.Slice(out,func(i,j int)bool{return out[i].Descriptor.RootCID<out[j].Descriptor.RootCID})
	return out
}

func (g *GestorRetencion) persistirLocked() error {
	if g.ruta=="" { return nil }
	b,err:=json.MarshalIndent(g.registros,"","  "); if err!=nil{return err}
	dir:=filepath.Dir(g.ruta)
	if err:=os.MkdirAll(dir,0700);err!=nil{return err}
	tmp,err:=os.CreateTemp(dir,".retenciones-*");if err!=nil{return err}
	nombre:=tmp.Name();ok:=false
	defer func(){if !ok{_ = os.Remove(nombre)}}()
	if err:=tmp.Chmod(0600);err!=nil{_ = tmp.Close();return err}
	if _,err:=tmp.Write(b);err!=nil{_ = tmp.Close();return err}
	if err:=tmp.Sync();err!=nil{_ = tmp.Close();return err}
	if err:=tmp.Close();err!=nil{return err}
	if err:=os.Rename(nombre,g.ruta);err!=nil{return err}
	if d,err:=os.Open(dir);err==nil{_ = d.Sync();_ = d.Close()}
	ok=true
	return nil
}

// FijarVersion registra retención local solo para una versión conocida y
// cuyo bloque está íntegro en el almacén.
func (n *Nodo) FijarVersion(d DescriptorVersion,factor int) error {
	if n==nil||n.almacen==nil||n.versiones==nil||n.retenciones==nil{return fmt.Errorf("servicios de retención no inicializados")}
	if err:=VerificarDescriptor(d);err!=nil{return err}
	conocida:=false
	for _,v:=range n.versiones.Versiones(d.OrganismoID){if v.RootCID==d.RootCID&&v.Firma==d.Firma{conocida=true;break}}
	if !conocida{return fmt.Errorf("el descriptor no está registrado en el índice local")}
	disponible,err:=n.almacen.Disponible(d.RootCID)
	if err!=nil||!disponible{return fmt.Errorf("no se puede fijar contenido ausente o corrupto: %w",err)}
	return n.retenciones.FijarVersion(d,factor,time.Now().UTC())
}

func (n *Nodo) LiberarVersion(root string) error {
	if n==nil||n.retenciones==nil{return fmt.Errorf("gestor de retención no inicializado")}
	return n.retenciones.Liberar(root)
}

func (n *Nodo) Retenciones() []RetencionVersion {
	if n==nil||n.retenciones==nil{return nil}
	return n.retenciones.Todas()
}

// RepararRetenciones comprueba cada pin local, recupera bloques ausentes o
// corruptos desde peers conectados/DHT y vuelve a asegurar su factor.
func (n *Nodo) RepararRetenciones(ctx context.Context) ([]EstadoRedundancia,error) {
	if n==nil||n.almacen==nil||n.retenciones==nil{return nil,fmt.Errorf("servicios de retención no inicializados")}
	var estados []EstadoRedundancia
	var errores []string
	for _,r:=range n.retenciones.Todas(){
		root:=r.Descriptor.RootCID
		contenido,err:=n.almacen.Obtener(root)
		if err!=nil{
			contenido,err=n.recuperarContenidoParaReparacion(ctx,root)
			if err==nil{err=n.almacen.Reparar(root,contenido)}
			if err!=nil{errores=append(errores,fmt.Sprintf("%s: recuperación local: %v",root,err));continue}
		}
		estado,err:=n.AsegurarRedundancia(ctx,r.Descriptor,contenido,r.Factor)
		estados=append(estados,estado)
		if err!=nil{errores=append(errores,fmt.Sprintf("%s: %v",root,err))}
	}
	if len(errores)>0{return estados,fmt.Errorf("reparación incompleta: %v",errores)}
	return estados,nil
}

func (n *Nodo) recuperarContenidoParaReparacion(ctx context.Context,root string)([]byte,error){
	peers:=n.Host.Network().Peers()
	sort.Slice(peers,func(i,j int)bool{return peers[i].String()<peers[j].String()})
	for _,p:=range peers{
		b,err:=n.leerBloquePeer(ctx,p,root);if err==nil{return b,nil}
	}
	if n.DHT==nil{return nil,fmt.Errorf("sin peers útiles ni DHT para recuperar %s",root)}
	clave,err:=CIDRoot(root);if err!=nil{return nil,err}
	proveedores,err:=n.DHT.FindProviders(ctx,clave);if err!=nil{return nil,err}
	for _,p:=range proveedores{
		if p.ID==n.ID(){continue}
		if err:=n.Host.Connect(ctx,p);err!=nil{continue}
		b,err:=n.leerBloquePeer(ctx,p.ID,root);if err==nil{return b,nil}
	}
	return nil,fmt.Errorf("ningún proveedor entregó contenido íntegro para %s",root)
}

func (n *Nodo) leerBloquePeer(ctx context.Context,p peer.ID,root string)([]byte,error){
	stream,err:=n.Host.NewStream(ctx,p,ProtocoloOrganismo);if err!=nil{return nil,err}
	defer stream.Close()
	if err:=json.NewEncoder(stream).Encode(solicitudOrganismo{Operacion:"obtener",RootCID:root});err!=nil{return nil,err}
	limitado:=io.LimitedReader{R:stream,N:maxContenidoVersion+(1<<20)}
	var respuesta respuestaOrganismo
	if err:=json.NewDecoder(&limitado).Decode(&respuesta);err!=nil{return nil,err}
	if respuesta.Error!=""||respuesta.RootCID!=root||limitado.N<=0{return nil,fmt.Errorf("peer no entregó el RootCID solicitado")}
	if len(respuesta.Manifiesto)>maxContenidoVersion{return nil,fmt.Errorf("bloque excede límite")}
	if err:=rootcid.Verificar(respuesta.Manifiesto,root);err!=nil{return nil,err}
	return append([]byte(nil),respuesta.Manifiesto...),nil
}

// IniciarMantenimiento activa reparación periódica de forma opt-in. El intervalo
// mínimo evita bucles agresivos; Cerrar detiene y espera el worker.
func (n *Nodo) IniciarMantenimiento(intervalo time.Duration) error {
	if n==nil||n.retenciones==nil{return fmt.Errorf("gestor de retención no inicializado")}
	if intervalo<time.Minute{return fmt.Errorf("intervalo de mantenimiento mínimo: 1 minuto")}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.cancelarMantenimiento!=nil{return fmt.Errorf("mantenimiento ya iniciado")}
	ctx,cancelar:=context.WithCancel(context.Background())
	hecho:=make(chan struct{})
	n.cancelarMantenimiento=cancelar
	n.mantenimientoDone=hecho
	go func(){
		defer close(hecho)
		t:=time.NewTicker(intervalo);defer t.Stop()
		for{
			select{
			case <-ctx.Done():return
			case <-t.C:
				limite,cancel:=context.WithTimeout(ctx,30*time.Second)
				if _,err:=n.RepararRetenciones(limite);err!=nil{fmt.Printf("[RETENTION-REPAIR-DEFERRED] %v\n",err)}
				cancel()
			}
		}
	}()
	return nil
}
