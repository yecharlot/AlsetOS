package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	ProtocoloSincronizacionVersiones = "/alset/version-sync/1.0.0"
	schemaSolicitudSincronizacion = "alset.version-sync.request.v1"
	maxInventarioVersiones = 8 << 20
	maxDescriptoresInventario = 10000
)

type solicitudSincronizacionVersiones struct {
	Schema string `json:"schema"`
}

type respuestaSincronizacionVersiones struct {
	Schema string `json:"schema"`
	Descriptores []DescriptorVersion `json:"descriptores"`
	Error string `json:"error,omitempty"`
}

// recibirSolicitudSincronizacion entrega una instantánea del inventario local.
// No transfiere contenido ni declara disponible ningún RootCID.
func (nodo *Nodo) recibirSolicitudSincronizacion(stream network.Stream) {
	defer stream.Close()
	if nodo == nil || nodo.versiones == nil {
		_ = json.NewEncoder(stream).Encode(respuestaSincronizacionVersiones{Error:"índice de versiones no disponible"})
		return
	}
	var solicitud solicitudSincronizacionVersiones
	limitado := &io.LimitedReader{R:stream, N:4096}
	if err := json.NewDecoder(bufio.NewReader(limitado)).Decode(&solicitud); err != nil || solicitud.Schema != schemaSolicitudSincronizacion {
		_ = json.NewEncoder(stream).Encode(respuestaSincronizacionVersiones{Error:"solicitud de sincronización inválida"})
		return
	}
	descriptores := nodo.versiones.Todas()
	if len(descriptores) > maxDescriptoresInventario {
		_ = json.NewEncoder(stream).Encode(respuestaSincronizacionVersiones{Error:"inventario supera el límite de descriptores"})
		return
	}
	respuesta := respuestaSincronizacionVersiones{Schema:"alset.version-sync.response.v1", Descriptores:descriptores}
	bytes, err := json.Marshal(respuesta)
	if err != nil || len(bytes) > maxInventarioVersiones {
		_ = json.NewEncoder(stream).Encode(respuestaSincronizacionVersiones{Error:"inventario supera el límite de transferencia"})
		return
	}
	_, _ = stream.Write(append(bytes, '\\n'))
}

// SincronizarVersiones solicita el inventario firmado del peer conectado,
// incorpora descriptores faltantes y encola la recuperación de sus RootCID.
// No resuelve conflictos ni sobrescribe versiones.
func (nodo *Nodo) SincronizarVersiones(ctx context.Context, remoto peer.ID) error {
	if nodo == nil || nodo.Host == nil || nodo.versiones == nil || nodo.almacen == nil {
		return fmt.Errorf("servicios de sincronización no inicializados")
	}
	if remoto == "" || remoto == nodo.ID() {
		return fmt.Errorf("peer remoto inválido")
	}
	stream, err := nodo.Host.NewStream(ctx, remoto, ProtocoloSincronizacionVersiones)
	if err != nil {
		return fmt.Errorf("abrir sincronización con %s: %w", remoto, err)
	}
	defer stream.Close()
	if err := json.NewEncoder(stream).Encode(solicitudSincronizacionVersiones{Schema:schemaSolicitudSincronizacion}); err != nil {
		return fmt.Errorf("solicitar inventario a %s: %w", remoto, err)
	}
	limitado := &io.LimitedReader{R:stream, N:maxInventarioVersiones+1}
	var respuesta respuestaSincronizacionVersiones
	if err := json.NewDecoder(bufio.NewReader(limitado)).Decode(&respuesta); err != nil {
		return fmt.Errorf("leer inventario de %s: %w", remoto, err)
	}
	if limitado.N <= 0 || respuesta.Schema != "alset.version-sync.response.v1" || respuesta.Error != "" {
		return fmt.Errorf("respuesta de inventario inválida de %s: %s", remoto, respuesta.Error)
	}
	if len(respuesta.Descriptores) > maxDescriptoresInventario {
		return fmt.Errorf("inventario remoto excede el límite")
	}
	for _, descriptor := range respuesta.Descriptores {
		if err := VerificarDescriptor(descriptor); err != nil {
			// Un descriptor inválido no bloquea el procesamiento de los demás.
			fmt.Printf("[VERSION-SYNC-REJECTED] peer=%s organismo=%s root=%s: %v\\n", remoto, descriptor.OrganismoID, descriptor.RootCID, err)
			continue
		}
		if _, err := nodo.versiones.Registrar(descriptor); err != nil {
			return fmt.Errorf("registrar descriptor %s/%s: %w", descriptor.OrganismoID, descriptor.RootCID, err)
		}
		select {
		case nodo.colaRecuperacion <- solicitudRecuperacionVersion{descriptor:descriptor, origen:remoto}:
		default:
			fmt.Printf("[VERSION-SYNC-QUEUE-FULL] peer=%s organismo=%s root=%s\\n", remoto, descriptor.OrganismoID, descriptor.RootCID)
		}
	}
	return nil
}

// ConnectedF inicia sincronización en segundo plano cada vez que un peer se
// conecta, incluido el retorno después de una partición de red.
func (nodo *Nodo) iniciarSincronizacionReconexiones() {
	if nodo == nil || nodo.Host == nil { return }
	observador := &network.NotifyBundle{
		ConnectedF: func(_ network.Network, conexion network.Conn) {
			remoto := conexion.RemotePeer()
			go func() {
				ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancelar()
				if err := nodo.SincronizarVersiones(ctx, remoto); err != nil {
					fmt.Printf("[VERSION-SYNC-ERROR] peer=%s: %v\\n", remoto, err)
				}
			}()
		},
	}
	nodo.Host.Network().Notify(observador)
	nodo.cancelarNotificaciones = func() { nodo.Host.Network().StopNotify(observador) }
}
