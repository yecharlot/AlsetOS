package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	datastore "github.com/ipfs/go-datastore"
	syncds "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/pulso"
	"github.com/yecharlot/AlsetOS/red"
)

const ProtocoloPulse = "/alset/pulse/1.0.0"

type Nodo struct {
	Host       host.Host
	Identidad  *identidad.Identidad
	DHT        *kaddht.IpfsDHT
	datastore  datastore.Batching
	mu         sync.RWMutex
	conocidos  map[peer.ID]peer.AddrInfo
	organismos map[string][]byte
	rutaOrganismos string
}

type descubrimiento struct { nodo *Nodo }

func (d *descubrimiento) HandlePeerFound(info peer.AddrInfo) {
	if info.ID == d.nodo.Host.ID() { return }
	d.nodo.mu.Lock()
	d.nodo.conocidos[info.ID] = info
	d.nodo.mu.Unlock()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = d.nodo.Host.Connect(ctx, info)
	}()
}

func Nuevo(identidadNodo *identidad.Identidad, escuchar string) (*Nodo, error) {
	return NuevoConEstado(identidadNodo, escuchar, "")
}

func NuevoConEstado(identidadNodo *identidad.Identidad, escuchar, rutaOrganismos string) (*Nodo, error) {
	if identidadNodo == nil { return nil, fmt.Errorf("identidad nula") }

	clave, err := libp2pcrypto.UnmarshalEd25519PrivateKey(identidadNodo.ClavePrivada)
	if err != nil { return nil, fmt.Errorf("convertir identidad Ed25519 a libp2p: %w", err) }

	h, err := libp2p.New(
		libp2p.Identity(clave),
		libp2p.ListenAddrStrings(escuchar),
	)
	if err != nil { return nil, fmt.Errorf("crear nodo libp2p: %w", err) }

	organismos, err := cargarOrganismos(rutaOrganismos)
	if err != nil {
		_ = h.Close()
		return nil, err
	}

	nodo := &Nodo{
		Host: h,
		Identidad: identidadNodo,
		datastore: syncds.MutexWrap(datastore.NewMapDatastore()),
		conocidos: make(map[peer.ID]peer.AddrInfo),
		organismos: organismos,
		rutaOrganismos: rutaOrganismos,
	}

	h.SetStreamHandler(ProtocoloPulse, nodo.recibirPulso)
	h.SetStreamHandler(ProtocoloOrganismo, nodo.recibirOrganismo)

	servicio := mdns.NewMdnsService(h, "_alset._tcp", &descubrimiento{nodo: nodo})
	if err := servicio.Start(); err != nil {
		_ = h.Close()
		return nil, fmt.Errorf("iniciar descubrimiento mDNS: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := iniciarDHT(ctx, nodo); err != nil {
		_ = h.Close()
		return nil, err
	}

	return nodo, nil
}

func iniciarDHT(ctx context.Context, nodo *Nodo) error {
	if nodo.DHT != nil { return nil }
	dht, err := kaddht.New(nodo.Host, kaddht.Datastore(nodo.datastore))
	if err != nil { return fmt.Errorf("crear DHT: %w", err) }
	nodo.DHT = dht
	return nil
}

func (nodo *Nodo) ReanunciarOrganismos(ctx context.Context) error {
	if nodo.DHT == nil { return fmt.Errorf("DHT no inicializada") }
	nodo.mu.RLock()
	copia := make(map[string][]byte, len(nodo.organismos))
	for root, contenido := range nodo.organismos { copia[root] = append([]byte(nil), contenido...) }
	nodo.mu.RUnlock()
	for root, contenido := range copia {
		clave, err := CIDRoot(root)
		if err != nil { return err }
		if err := nodo.DHT.Provide(ctx, clave, true); err != nil { return fmt.Errorf("reanunciar %s: %w", root, err) }
		_ = contenido
	}
	return nil
}

func (nodo *Nodo) BootstrapDHT(ctx context.Context) error {
	if nodo.DHT == nil { return fmt.Errorf("DHT no inicializada") }

	for _, peerID := range nodo.Host.Network().Peers() {
		if err := nodo.DHT.Ping(ctx, peerID); err == nil {
			if _, err := nodo.DHT.RoutingTable().TryAddPeer(peerID, true, true); err != nil {
				return fmt.Errorf("añadir peer al routing table: %w", err)
			}
		}
	}

	if err := nodo.DHT.Bootstrap(ctx); err != nil {
		return fmt.Errorf("bootstrap DHT: %w", err)
	}
	return nil
}

func (nodo *Nodo) recibirPulso(stream network.Stream) {
	defer stream.Close()
	conexion := red.NuevaConexion(stream)
	mensaje, err := conexion.RecibirPulsoFirmado()
	if err != nil {
		fmt.Printf("[P2P-RECHAZADO] %v\n", err)
		return
	}
	fmt.Printf("[P2P-RECIBIDO] nodo=%s tipo=%s contenido=%s\n", mensaje.NodoID, mensaje.Pulso.Tipo, mensaje.Pulso.Contenido)
}

func (nodo *Nodo) EnviarPulso(ctx context.Context, destino peer.ID, evento pulso.Pulso) error {
	stream, err := nodo.Host.NewStream(ctx, destino, ProtocoloPulse)
	if err != nil { return fmt.Errorf("abrir stream Pulse: %w", err) }
	defer stream.Close()
	if err := red.NuevaConexion(stream).EnviarPulsoFirmado(nodo.Identidad, evento); err != nil {
		return fmt.Errorf("enviar Pulse: %w", err)
	}
	return nil
}

func (nodo *Nodo) Conocidos() []peer.AddrInfo {
	nodo.mu.RLock()
	defer nodo.mu.RUnlock()
	resultado := make([]peer.AddrInfo, 0, len(nodo.conocidos))
	for _, info := range nodo.conocidos { resultado = append(resultado, info) }
	return resultado
}

func (nodo *Nodo) ID() peer.ID { return nodo.Host.ID() }

func (nodo *Nodo) Direcciones() []string {
	resultado := make([]string, 0, len(nodo.Host.Addrs()))
	for _, direccion := range nodo.Host.Addrs() {
		resultado = append(resultado, fmt.Sprintf("%s/p2p/%s", direccion, nodo.ID()))
	}
	return resultado
}

func (nodo *Nodo) Cerrar() error {
	if nodo.DHT != nil {
		if err := nodo.DHT.Close(); err != nil { return err }
	}
	return nodo.Host.Close()
}
