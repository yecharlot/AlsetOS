package p2p

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	ProtocoloLeaseRetencion = "/alset/retention-lease/1.0.0"
	EsquemaLeaseRetencion = "alset.retention-lease.v1"
	maxDuracionLease = 30 * 24 * time.Hour
)

type LeaseRetencion struct {
	Schema string `json:"schema"`
	ID string `json:"id"`
	RootCID string `json:"root_cid"`
	OrganismoID string `json:"organismo_id"`
	Solicitante string `json:"solicitante"`
	Proveedor string `json:"proveedor"`
	Bytes int64 `json:"bytes"`
	Emitido time.Time `json:"emitido"`
	Expira time.Time `json:"expira"`
	ClavePublica string `json:"clave_publica"`
	Firma string `json:"firma"`
}

type leaseFirmable struct {
	Schema string `json:"schema"`
	ID string `json:"id"`
	RootCID string `json:"root_cid"`
	OrganismoID string `json:"organismo_id"`
	Solicitante string `json:"solicitante"`
	Proveedor string `json:"proveedor"`
	Bytes int64 `json:"bytes"`
	Emitido time.Time `json:"emitido"`
	Expira time.Time `json:"expira"`
	ClavePublica string `json:"clave_publica"`
}

func (l LeaseRetencion) firmable() leaseFirmable {
	return leaseFirmable{l.Schema,l.ID,l.RootCID,l.OrganismoID,l.Solicitante,l.Proveedor,l.Bytes,l.Emitido,l.Expira,l.ClavePublica}
}

type solicitudLease struct {
	Accion string `json:"accion"`
	Lease LeaseRetencion `json:"lease"`
	Descriptor DescriptorVersion `json:"descriptor"`
	ClavePublica string `json:"clave_publica"`
	Firma string `json:"firma"`
}

type solicitudLeaseFirmable struct {
	Accion string `json:"accion"`
	Lease LeaseRetencion `json:"lease"`
	Descriptor DescriptorVersion `json:"descriptor"`
	ClavePublica string `json:"clave_publica"`
}

type respuestaLease struct {
	Aceptada bool `json:"aceptada"`
	Motivo string `json:"motivo,omitempty"`
	Lease LeaseRetencion `json:"lease"`
}

type GestorLeases struct {
	mu sync.RWMutex
	ruta string
	leases map[string]LeaseRetencion
	permitidos map[string]int64
	capacidad int64
}

func NuevoGestorLeases(ruta string) (*GestorLeases,error) {
	g:=&GestorLeases{ruta:ruta,leases:map[string]LeaseRetencion{},permitidos:map[string]int64{}}
	if ruta=="" { return g,nil }
	b,err:=os.ReadFile(ruta)
	if os.IsNotExist(err) { return g,nil }
	if err!=nil { return nil,err }
	if len(b)==0 { return g,nil }
	if err:=json.Unmarshal(b,&g.leases);err!=nil{return nil,fmt.Errorf("leer leases de retención: %w",err)}
	for id,l:=range g.leases {
		if id!=l.ID || l.Schema!=EsquemaLeaseRetencion || l.Bytes<=0 || l.Expira.IsZero() || !verificarFirmaLease(l) {
			return nil,fmt.Errorf("lease persistido inválido: %s",id)
		}
	}
	return g,nil
}

// AutorizarPeer es una decisión local explícita. La autorización no se deriva
// de la firma del descriptor ni de la identidad criptográfica del solicitante.
func (g *GestorLeases) AutorizarPeer(id peer.ID, cuota int64) error {
	if g==nil || id=="" || cuota<=0{return fmt.Errorf("peer o cuota de retención inválidos")}
	g.mu.Lock();defer g.mu.Unlock()
	g.permitidos[id.String()]=cuota
	return nil
}
func (g *GestorLeases) RevocarPeer(id peer.ID) {
	if g==nil{return};g.mu.Lock();delete(g.permitidos,id.String());g.mu.Unlock()
}
func (g *GestorLeases) ConfigurarCapacidad(bytes int64) error {
	if g==nil||bytes<=0{return fmt.Errorf("capacidad de retención debe ser positiva")}
	g.mu.Lock();defer g.mu.Unlock()
	if g.usadosLocked(time.Now().UTC())>bytes{return fmt.Errorf("la capacidad nueva es menor que el almacenamiento comprometido")}
	g.capacidad=bytes;return nil
}
func (g *GestorLeases) usadosLocked(ahora time.Time) int64 {
	var total int64
	for _,l:=range g.leases{if l.Expira.After(ahora){total+=l.Bytes}}
	return total
}
func (g *GestorLeases) persistirLocked() error {
	if g.ruta==""{return nil}
	b,err:=json.MarshalIndent(g.leases,"","  ");if err!=nil{return err}
	dir:=filepath.Dir(g.ruta);if err:=os.MkdirAll(dir,0700);err!=nil{return err}
	f,err:=os.CreateTemp(dir,".leases-*");if err!=nil{return err}
	name:=f.Name();ok:=false;defer func(){if !ok{_ = os.Remove(name)}}()
	if err:=f.Chmod(0600);err!=nil{_ = f.Close();return err}
	if _,err:=f.Write(b);err!=nil{_ = f.Close();return err}
	if err:=f.Sync();err!=nil{_ = f.Close();return err}
	if err:=f.Close();err!=nil{return err}
	if err:=os.Rename(name,g.ruta);err!=nil{return err}
	if d,err:=os.Open(dir);err==nil{_ = d.Sync();_ = d.Close()}
	ok=true;return nil
}
func (g *GestorLeases) Activos(ahora time.Time) []LeaseRetencion {
	if g==nil{return nil};g.mu.RLock();defer g.mu.RUnlock()
	out:=[]LeaseRetencion{};for _,l:=range g.leases{if l.Expira.After(ahora){out=append(out,l)}}
	sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID});return out
}
func (g *GestorLeases) Expirar(ahora time.Time) error {
	if g==nil{return nil};g.mu.Lock();defer g.mu.Unlock()
	prev:=make(map[string]LeaseRetencion,len(g.leases));for k,v:=range g.leases{prev[k]=v}
	for id,l:=range g.leases{if !l.Expira.After(ahora){delete(g.leases,id)}}
	if err:=g.persistirLocked();err!=nil{g.leases=prev;return err};return nil
}
func verificarFirmaLease(l LeaseRetencion) bool {
	pub,e:=base64.StdEncoding.DecodeString(l.ClavePublica);if e!=nil||len(pub)!=ed25519.PublicKeySize{return false}
	firma,e:=base64.StdEncoding.DecodeString(l.Firma);if e!=nil||len(firma)!=ed25519.SignatureSize{return false}
	pid,e:=peer.Decode(l.Proveedor);if e!=nil||peerIDDeClave(ed25519.PublicKey(pub))!=pid.String(){return false}
	b,e:=json.Marshal(l.firmable());return e==nil&&ed25519.Verify(ed25519.PublicKey(pub),b,firma)
}
func verificarSolicitudLease(s solicitudLease, remoto peer.ID) error {
	pub,e:=base64.StdEncoding.DecodeString(s.ClavePublica);if e!=nil||len(pub)!=ed25519.PublicKeySize{return fmt.Errorf("clave del solicitante inválida")}
	if peerIDDeClave(ed25519.PublicKey(pub))!=remoto.String(){return fmt.Errorf("la clave del solicitante no corresponde al peer conectado")}
	sig,e:=base64.StdEncoding.DecodeString(s.Firma);if e!=nil||len(sig)!=ed25519.SignatureSize{return fmt.Errorf("firma de solicitud inválida")}
	b,e:=json.Marshal(solicitudLeaseFirmable{s.Accion,s.Lease,s.Descriptor,s.ClavePublica});if e!=nil{return e}
	if !ed25519.Verify(ed25519.PublicKey(pub),b,sig){return fmt.Errorf("firma de solicitud no válida")}
	return nil
}
func (n *Nodo) AutorizarRetencionRemota(id peer.ID, cuotaBytes int64) error {
	if n==nil||n.leases==nil{return fmt.Errorf("gestor de leases no inicializado")}
	return n.leases.AutorizarPeer(id,cuotaBytes)
}
func (n *Nodo) RevocarRetencionRemota(id peer.ID) { if n!=nil&&n.leases!=nil{n.leases.RevocarPeer(id)} }
func (n *Nodo) ConfigurarCapacidadRetencion(bytes int64) error {
	if n==nil||n.leases==nil{return fmt.Errorf("gestor de leases no inicializado")}
	return n.leases.ConfigurarCapacidad(bytes)
}
func (n *Nodo) LeasesRetencion() []LeaseRetencion { if n==nil||n.leases==nil{return nil};return n.leases.Activos(time.Now().UTC()) }

// SolicitarRetencionRemota presupone que el contenido ya fue replicado y
// verificado en el peer. La aceptación firmada constituye un compromiso con
// vencimiento, no una garantía física futura.
func (n *Nodo) SolicitarRetencionRemota(ctx context.Context, destino peer.ID, d DescriptorVersion, duracion time.Duration) (LeaseRetencion,error) {
	if n==nil||n.Host==nil||n.Identidad==nil{return LeaseRetencion{},fmt.Errorf("nodo no inicializado")}
	if err:=VerificarDescriptor(d);err!=nil{return LeaseRetencion{},err}
	if duracion<=0||duracion>maxDuracionLease{return LeaseRetencion{},fmt.Errorf("duración de lease fuera de rango")}
	if _,err:=n.almacen.Obtener(d.RootCID);err!=nil{return LeaseRetencion{},fmt.Errorf("coordinador no conserva el bloque íntegro: %w",err)}
	now:=time.Now().UTC();nonce:=make([]byte,24);if _,err:=rand.Read(nonce);err!=nil{return LeaseRetencion{},err}
	l:=LeaseRetencion{Schema:EsquemaLeaseRetencion,ID:base64.RawURLEncoding.EncodeToString(nonce),RootCID:d.RootCID,OrganismoID:d.OrganismoID,Solicitante:n.ID().String(),Proveedor:destino.String(),Bytes:0,Emitido:now,Expira:now.Add(duracion)}
	l.ClavePublica=base64.StdEncoding.EncodeToString(n.Identidad.ClavePublica)
	s:=solicitudLease{Accion:"aceptar",Lease:l,Descriptor:d,ClavePublica:base64.StdEncoding.EncodeToString(n.Identidad.ClavePublica)}
	payload,_:=json.Marshal(solicitudLeaseFirmable{s.Accion,s.Lease,s.Descriptor,s.ClavePublica});s.Firma=base64.StdEncoding.EncodeToString(n.Identidad.Firmar(payload))
	return n.enviarSolicitudLease(ctx,destino,s)
}
func (n *Nodo) RenovarLeaseRemoto(ctx context.Context, lease LeaseRetencion, duracion time.Duration) (LeaseRetencion,error) {
	if duracion<=0||duracion>maxDuracionLease{return LeaseRetencion{},fmt.Errorf("duración de lease fuera de rango")}
	lease.Emitido=time.Now().UTC();lease.Expira=lease.Emitido.Add(duracion);lease.Firma=""
	s:=solicitudLease{Accion:"renovar",Lease:lease,ClavePublica:base64.StdEncoding.EncodeToString(n.Identidad.ClavePublica)}
	payload,_:=json.Marshal(solicitudLeaseFirmable{s.Accion,s.Lease,DescriptorVersion{},s.ClavePublica});s.Firma=base64.StdEncoding.EncodeToString(n.Identidad.Firmar(payload))
	return n.enviarSolicitudLease(ctx,peer.ID(lease.Proveedor),s)
}
// LiberarLeaseRemoto termina el compromiso contractual sin borrar el bloque.
func (n *Nodo) LiberarLeaseRemoto(ctx context.Context, lease LeaseRetencion) error {
	if n==nil||n.Identidad==nil{return fmt.Errorf("nodo no inicializado")}
	lease.Firma=""
	s:=solicitudLease{Accion:"liberar",Lease:lease,ClavePublica:base64.StdEncoding.EncodeToString(n.Identidad.ClavePublica)}
	payload,_:=json.Marshal(solicitudLeaseFirmable{s.Accion,s.Lease,DescriptorVersion{},s.ClavePublica})
	s.Firma=base64.StdEncoding.EncodeToString(n.Identidad.Firmar(payload))
	_,err:=n.enviarSolicitudLease(ctx,peer.ID(lease.Proveedor),s)
	return err
}

func (n *Nodo) enviarSolicitudLease(ctx context.Context,destino peer.ID,s solicitudLease)(LeaseRetencion,error){
	stream,err:=n.Host.NewStream(ctx,destino,ProtocoloLeaseRetencion);if err!=nil{return LeaseRetencion{},err};defer stream.Close()
	if err:=json.NewEncoder(stream).Encode(s);err!=nil{return LeaseRetencion{},err}
	lim:=&io.LimitedReader{R:stream,N:32<<10};var r respuestaLease
	if err:=json.NewDecoder(bufio.NewReader(lim)).Decode(&r);err!=nil{return LeaseRetencion{},err}
	if lim.N<=0{return LeaseRetencion{},fmt.Errorf("respuesta de lease excede límite")}
	if !r.Aceptada{return LeaseRetencion{},fmt.Errorf("lease rechazado: %s",r.Motivo)}
	if s.Accion=="liberar" { return r.Lease,nil }
	if r.Lease.ID!=s.Lease.ID||r.Lease.RootCID!=s.Lease.RootCID||r.Lease.Solicitante!=n.ID().String()||r.Lease.Proveedor!=destino.String()||!r.Lease.Expira.After(time.Now().UTC())||!verificarFirmaLease(r.Lease){return LeaseRetencion{},fmt.Errorf("compromiso remoto inválido")}
	return r.Lease,nil
}
func (n *Nodo) recibirSolicitudLease(stream network.Stream) {
	defer stream.Close()
	if n==nil||n.leases==nil||n.Identidad==nil||n.almacen==nil{_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"servicio no disponible"});return}
	lim:=&io.LimitedReader{R:stream,N:64<<10};var s solicitudLease
	if err:=json.NewDecoder(bufio.NewReader(lim)).Decode(&s);err!=nil||lim.N<=0{_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"solicitud inválida"});return}
	remoto:=stream.Conn().RemotePeer()
	if err:=verificarSolicitudLease(s,remoto);err!=nil{_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:err.Error()});return}
	if s.Accion!="aceptar"&&s.Accion!="renovar"&&s.Accion!="liberar"{_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"acción no soportada"});return}
	l:=s.Lease
	if l.Schema!=EsquemaLeaseRetencion||l.Solicitante!=remoto.String()||l.Proveedor!=n.ID().String()||l.ID==""||l.Expira.After(time.Now().UTC().Add(maxDuracionLease))||(s.Accion!="liberar"&&!l.Expira.After(time.Now().UTC())){_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"parámetros de lease inválidos"});return}
	if s.Accion=="aceptar" {
		if err:=VerificarDescriptor(s.Descriptor);err!=nil||s.Descriptor.RootCID!=l.RootCID||s.Descriptor.OrganismoID!=l.OrganismoID{_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"descriptor no corresponde al lease"});return}
	}
	if s.Accion=="liberar" {
		n.leases.mu.Lock()
		prev,ok:=n.leases.leases[l.ID]
		if !ok || prev.Solicitante!=remoto.String() || prev.Proveedor!=n.ID().String() || prev.RootCID!=l.RootCID {
			n.leases.mu.Unlock();_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"lease desconocido o no corresponde al solicitante"});return
		}
		delete(n.leases.leases,l.ID)
		if err:=n.leases.persistirLocked();err!=nil{n.leases.leases[l.ID]=prev;n.leases.mu.Unlock();_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"no se pudo persistir la liberación"});return}
		n.leases.mu.Unlock()
		_ = json.NewEncoder(stream).Encode(respuestaLease{Aceptada:true,Lease:prev})
		return
	}
	bytes,err:=n.almacen.Obtener(l.RootCID);if err!=nil{_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"bloque ausente o corrupto"});return}
	l.Bytes=int64(len(bytes));l.Emitido=time.Now().UTC()
	n.leases.mu.Lock()
	cuota,ok:=n.leases.permitidos[remoto.String()]
	if !ok{n.leases.mu.Unlock();_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"peer no autorizado para solicitar retención"});return}
	prev,existe:=n.leases.leases[l.ID]
	if s.Accion=="renovar"&&!existe{n.leases.mu.Unlock();_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"lease desconocido"});return}
	if s.Accion=="aceptar"&&existe{n.leases.mu.Unlock();_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"ID de lease ya existente"});return}
	if s.Accion=="renovar"&&(l.RootCID!=prev.RootCID||l.OrganismoID!=prev.OrganismoID||l.Solicitante!=prev.Solicitante||l.Proveedor!=prev.Proveedor){n.leases.mu.Unlock();_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"la renovación altera la identidad del compromiso"});return}
	var usadosPeer int64
	for id,x:=range n.leases.leases{if id!=l.ID&&x.Solicitante==remoto.String()&&x.Expira.After(time.Now().UTC()){usadosPeer+=x.Bytes}}
	if l.Bytes>cuota-usadosPeer{n.leases.mu.Unlock();_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"cuota del solicitante insuficiente"});return}
	usados:=n.leases.usadosLocked(time.Now().UTC())
	if existe && prev.Expira.After(time.Now().UTC()) { usados-=prev.Bytes }
	if n.leases.capacidad==0 || usados+l.Bytes>n.leases.capacidad { n.leases.mu.Unlock();_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"capacidad de almacenamiento no configurada o insuficiente"});return }
	l.ClavePublica=base64.StdEncoding.EncodeToString(n.Identidad.ClavePublica)
	payload,_:=json.Marshal(l.firmable());l.Firma=base64.StdEncoding.EncodeToString(n.Identidad.Firmar(payload))
	n.leases.leases[l.ID]=l
	if err:=n.leases.persistirLocked();err!=nil{if existe{n.leases.leases[l.ID]=prev}else{delete(n.leases.leases,l.ID)};n.leases.mu.Unlock();_ = json.NewEncoder(stream).Encode(respuestaLease{Motivo:"no se pudo persistir compromiso"});return}
	n.leases.mu.Unlock()
	_ = json.NewEncoder(stream).Encode(respuestaLease{Aceptada:true,Lease:l})
}


// ExpirarLeases elimina únicamente compromisos vencidos del registro. El bloque
// físico se conserva: esta operación nunca ejecuta garbage collection.
func (n *Nodo) ExpirarLeases(ahora time.Time) error {
	if n==nil||n.leases==nil{return fmt.Errorf("gestor de leases no inicializado")}
	return n.leases.Expirar(ahora.UTC())
}
