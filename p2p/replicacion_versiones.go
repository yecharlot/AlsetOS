package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/yecharlot/AlsetOS/rootcid"
)

const ProtocoloVersion = "/alset/version/1.0.0"

type solicitudVersion struct {
	Descriptor DescriptorVersion `json:"descriptor"`
	Contenido []byte `json:"contenido"`
}

type respuestaVersion struct {
	RootCID string `json:"root_cid"`
	Estado EstadoLinaje `json:"estado"`
	Error string `json:"error,omitempty"`
}

// AnunciarVersion crea y firma una revisión con la identidad persistente del
// nodo. Si destino no está vacío, envía también el bloque y el descriptor.
// OrganismoID es una clave lógica del dominio; no se deriva del RootCID.
func (nodo *Nodo) AnunciarVersion(ctx context.Context, destino peer.ID, organismoID string, revision uint64, root, anterior string, contenido []byte) (DescriptorVersion, EstadoLinaje, error) {
	if nodo == nil || nodo.Identidad == nil || nodo.almacen == nil || nodo.versiones == nil {
		return DescriptorVersion{}, "", fmt.Errorf("nodo o servicios de replicación no inicializados")
	}
	if err := rootcid.Verificar(contenido, root); err != nil { return DescriptorVersion{}, "", err }
	descriptor, err := CrearDescriptorVersion(nodo.Identidad, organismoID, revision, root, anterior)
	if err != nil { return DescriptorVersion{}, "", err }
	if err := nodo.almacen.Guardar(root, contenido); err != nil { return DescriptorVersion{}, "", err }
	estado, err := nodo.versiones.Registrar(descriptor)
	if err != nil { return DescriptorVersion{}, "", err }
	if destino != "" {
		estadoRemoto, err := nodo.ReplicarDescriptor(ctx, destino, descriptor, contenido)
		if err != nil { return descriptor, estado, err }
		if estadoRemoto == LinajeConflicto { return descriptor, estadoRemoto, ErrLinajeEnConflicto }
		if estadoRemoto == LinajeIncompleto { return descriptor, estadoRemoto, ErrLinajeIncompleto }
	}
	return descriptor, estado, nil
}

// ReplicarDescriptor transfiere contenido inmutable y procedencia firmada.
// La firma prueba quién publicó el descriptor; no concede autorización para
// ejecutar el organismo ni convierte al replicador en autoridad de confianza.
func (nodo *Nodo) ReplicarDescriptor(ctx context.Context, destino peer.ID, descriptor DescriptorVersion, contenido []byte) (EstadoLinaje, error) {
	if nodo == nil || nodo.Host == nil { return "", fmt.Errorf("nodo no inicializado") }
	if err := VerificarDescriptor(descriptor); err != nil { return "", err }
	if err := rootcid.Verificar(contenido, descriptor.RootCID); err != nil { return "", err }
	stream, err := nodo.Host.NewStream(ctx, destino, ProtocoloVersion)
	if err != nil { return "", fmt.Errorf("abrir stream de versión: %w", err) }
	defer stream.Close()
	if err := json.NewEncoder(stream).Encode(solicitudVersion{Descriptor:descriptor, Contenido:contenido}); err != nil {
		return "", fmt.Errorf("enviar descriptor de versión: %w", err)
	}
	var respuesta respuestaVersion
	if err := json.NewDecoder(bufio.NewReader(stream)).Decode(&respuesta); err != nil {
		return "", fmt.Errorf("leer confirmación de versión: %w", err)
	}
	if respuesta.Error != "" { return respuesta.Estado, fmt.Errorf("versión rechazada: %s", respuesta.Error) }
	if respuesta.RootCID != descriptor.RootCID { return "", fmt.Errorf("confirmación de versión no corresponde al RootCID") }
	return respuesta.Estado, nil
}

func (nodo *Nodo) recibirVersion(stream network.Stream) {
	defer stream.Close()
	if nodo == nil || nodo.almacen == nil || nodo.versiones == nil {
		_ = json.NewEncoder(stream).Encode(respuestaVersion{Error:"servicio de versiones no disponible"})
		return
	}
	var solicitud solicitudVersion
	if err := json.NewDecoder(bufio.NewReader(stream)).Decode(&solicitud); err != nil {
		_ = json.NewEncoder(stream).Encode(respuestaVersion{Error:"solicitud de versión inválida"})
		return
	}
	d := solicitud.Descriptor
	if err := VerificarDescriptor(d); err != nil {
		_ = json.NewEncoder(stream).Encode(respuestaVersion{RootCID:d.RootCID, Error:err.Error()})
		return
	}
	if err := rootcid.Verificar(solicitud.Contenido, d.RootCID); err != nil {
		_ = json.NewEncoder(stream).Encode(respuestaVersion{RootCID:d.RootCID, Error:err.Error()})
		return
	}
	// El bloque es inmutable y verificable; después se registra la procedencia.
	// Si falla el índice, el bloque queda como contenido huérfano recuperable,
	// pero nunca se anuncia como una versión aceptada.
	if err := nodo.almacen.Guardar(d.RootCID, solicitud.Contenido); err != nil {
		_ = json.NewEncoder(stream).Encode(respuestaVersion{RootCID:d.RootCID, Error:err.Error()})
		return
	}
	estado, err := nodo.versiones.Registrar(d)
	if err != nil {
		_ = json.NewEncoder(stream).Encode(respuestaVersion{RootCID:d.RootCID, Error:err.Error()})
		return
	}
	_ = json.NewEncoder(stream).Encode(respuestaVersion{RootCID:d.RootCID, Estado:estado})
}

func (nodo *Nodo) EstadoLinaje(organismoID string) EstadoLinaje {
	if nodo == nil || nodo.versiones == nil { return LinajeIncompleto }
	return nodo.versiones.Estado(organismoID)
}

func (nodo *Nodo) Versiones(organismoID string) []DescriptorVersion {
	if nodo == nil || nodo.versiones == nil { return nil }
	return nodo.versiones.Versiones(organismoID)
}


// ResultadoReplicacion describe la confirmación individual de un peer.
// Un resultado no equivale a autorización ni a garantía de retención futura.
type ResultadoReplicacion struct {
	PeerID string `json:"peer_id"`
	Estado EstadoLinaje `json:"estado"`
	Error string `json:"error,omitempty"`
}

// ReplicarVersionAConocidos realiza fan-out explícito a los peers actualmente
// conectados. No descubre peers, no crea una política automática de retención
// y no afirma que el bloque siga disponible después de la confirmación.
func (nodo *Nodo) ReplicarVersionAConocidos(ctx context.Context, descriptor DescriptorVersion, contenido []byte) []ResultadoReplicacion {
	if nodo == nil || nodo.Host == nil { return []ResultadoReplicacion{{Error:"nodo no inicializado"}} }
	peers := nodo.Host.Network().Peers()
	sort.Slice(peers, func(i,j int) bool { return peers[i].String() < peers[j].String() })
	resultados := make([]ResultadoReplicacion, 0, len(peers))
	for _, destino := range peers {
		estado, err := nodo.ReplicarDescriptor(ctx, destino, descriptor, contenido)
		resultado := ResultadoReplicacion{PeerID:destino.String(), Estado:estado}
		if err != nil { resultado.Error = err.Error() }
		resultados = append(resultados, resultado)
	}
	return resultados
}
