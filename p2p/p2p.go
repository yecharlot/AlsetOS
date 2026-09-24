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
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/pulso"
	"github.com/yecharlot/AlsetOS/red"
)

const ProtocoloPulse = "/alset/pulse/1.0.0"

type Nodo struct {
	Host host.Host
	Identidad *identidad.Identidad
	mu sync.Mutex
	conocidos map[peer.ID]peer.AddrInfo
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
	if identidadNodo == nil { return nil, fmt.Errorf("identidad nula") }
	clave, err := libp2pcrypto.UnmarshalEd25519PrivateKey(identidadNodo.ClavePrivada)
	if err != nil { return nil, fmt.Errorf("convertir identidad Ed25519 a libp2p: %w", err) }

	h, err := libp2p.New(
		libp2p.Identity(clave),
		libp2p.ListenAddrStrings(escuchar),
	)
	if err != nil { return nil, fmt.Errorf("crear nodo libp2p: %w", err) }

	nodo := &Nodo{Host:h, Identidad:identidadNodo, conocidos:make(map[peer.ID]peer.AddrInfo)}
	h.SetStreamHandler(ProtocoloPulse, nodo.recibirPulso)

	servicio := mdns.NewMdnsService(h, "_alset._tcp", &descubrimiento{nodo:nodo})
	if err := servicio.Start(); err != nil {
		_ = h.Close()
		return nil, fmt.Errorf("iniciar descubrimiento mDNS: %w", err)
	}
	return nodo, nil
}

func (nodo *Nodo) recibirPulso(stream network.Stream) {
	defer stream.Close()
	conexion := red.NuevaConexion(stream)
	mensaje, err := conexion.RecibirPulsoFirmado()
	if err != nil { return }
	fmt.Printf("[P2P-RECIBIDO] nodo=%s tipo=%s contenido=%s\n",
		mensaje.NodoID, mensaje.Pulso.Tipo, mensaje.Pulso.Contenido)
}

func (nodo *Nodo) EnviarPulso(ctx context.Context, destino peer.ID, evento pulso.Pulso) error {
	stream, err := nodo.Host.NewStream(ctx, destino, ProtocoloPulse)
	if err != nil { return fmt.Errorf("abrir stream Pulse: %w", err) }
	defer stream.Close()
	return red.NuevaConexion(stream).EnviarPulsoFirmado(nodo.Identidad, evento)
}

func (nodo *Nodo) Conocidos() []peer.AddrInfo {
	nodo.mu.Lock()
	defer nodo.mu.Unlock()
	resultado := make([]peer.AddrInfo, 0, len(nodo.conocidos))
	for _, info := range nodo.conocidos { resultado = append(resultado, info) }
	return resultado
}

func (nodo *Nodo) ID() peer.ID { return nodo.Host.ID() }
func (nodo *Nodo) Cerrar() error { return nodo.Host.Close() }
