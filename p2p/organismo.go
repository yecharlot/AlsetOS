package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/yecharlot/AlsetOS/rootcid"
)

const ProtocoloOrganismo = "/alset/organism/1.0.0"

type solicitudOrganismo struct {
	Operacion  string
	RootCID    string
	Manifiesto []byte
}

type respuestaOrganismo struct {
	RootCID    string
	Manifiesto []byte
	Error      string
}

func normalizarRootCID(valor string) string {
	return strings.TrimSpace(valor)
}

func (nodo *Nodo) AnunciarOrganismo(ctx context.Context, rootCID string, manifiesto []byte) error {
	if nodo == nil { return fmt.Errorf("nodo nulo") }
	if nodo.DHT == nil { return fmt.Errorf("DHT no inicializada") }
	rootCID = normalizarRootCID(rootCID)
	if rootCID == "" {
		return fmt.Errorf("RootCID vacío")
	}
	if rootcid.CrearContenido(manifiesto) != rootCID {
		return fmt.Errorf("el manifiesto no corresponde al RootCID")
	}

	if nodo.almacen == nil { return fmt.Errorf("almacén de contenido no inicializado") }
	if err := nodo.almacen.Guardar(rootCID, manifiesto); err != nil { return fmt.Errorf("guardar bloque verificado: %w", err) }

	nodo.mu.Lock()
	copia := make(map[string][]byte, len(nodo.organismos)+1)
	for root, contenido := range nodo.organismos { copia[root] = append([]byte(nil), contenido...) }
	copia[rootCID] = append([]byte(nil), manifiesto...)
	if err := guardarOrganismos(nodo.rutaOrganismos, copia); err != nil {
		nodo.mu.Unlock()
		return err
	}
	nodo.organismos = copia
	nodo.mu.Unlock()

	clave, err := CIDRoot(rootCID)
	if err != nil {
		return err
	}
	if err := nodo.DHT.Provide(ctx, clave, true); err != nil {
		return fmt.Errorf("anunciar RootCID en DHT: %w", err)
	}
	return nil
}

func (nodo *Nodo) ReplicarOrganismo(ctx context.Context, destino peer.ID, rootCID string, manifiesto []byte) error {
	rootCID = normalizarRootCID(rootCID)
	if rootcid.CrearContenido(manifiesto) != rootCID {
		return fmt.Errorf("manifiesto y RootCID no coinciden")
	}

	stream, err := nodo.Host.NewStream(ctx, destino, ProtocoloOrganismo)
	if err != nil {
		return fmt.Errorf("abrir stream de réplica: %w", err)
	}
	defer stream.Close()

	if err := json.NewEncoder(stream).Encode(solicitudOrganismo{
		Operacion:  "replicar",
		RootCID:    rootCID,
		Manifiesto: manifiesto,
	}); err != nil {
		return fmt.Errorf("enviar réplica: %w", err)
	}

	var respuesta respuestaOrganismo
	if err := json.NewDecoder(bufio.NewReader(stream)).Decode(&respuesta); err != nil {
		return fmt.Errorf("leer confirmación de réplica: %w", err)
	}
	if respuesta.Error != "" {
		return fmt.Errorf("réplica rechazada: %s", respuesta.Error)
	}
	if respuesta.RootCID != rootCID {
		return fmt.Errorf("confirmación de réplica inválida")
	}
	return nil
}

func (nodo *Nodo) buscarProveedor(ctx context.Context, rootCID string) ([]byte, error) {
	clave, err := CIDRoot(rootCID)
	if err != nil {
		return nil, err
	}
	proveedores, err := nodo.DHT.FindProviders(ctx, clave)
	if err != nil {
		return nil, fmt.Errorf("buscar proveedor RootCID: %w", err)
	}
	for _, proveedor := range proveedores {
		if proveedor.ID == nodo.ID() {
			if local, err := nodo.almacen.Obtener(rootCID); err == nil { return local, nil }
		}
		if err := nodo.Host.Connect(ctx, proveedor); err != nil {
			continue
		}
		stream, err := nodo.Host.NewStream(ctx, proveedor.ID, ProtocoloOrganismo)
		if err != nil {
			continue
		}
		_ = json.NewEncoder(stream).Encode(solicitudOrganismo{
			Operacion: "obtener",
			RootCID:   rootCID,
		})
		var respuesta respuestaOrganismo
		err = json.NewDecoder(bufio.NewReader(stream)).Decode(&respuesta)
		_ = stream.Close()
		if err != nil {
			continue
		}
		if respuesta.RootCID != rootCID {
			continue
		}
		if rootcid.CrearContenido(respuesta.Manifiesto) != rootCID {
			continue
		}
		if err := nodo.almacen.Guardar(rootCID, respuesta.Manifiesto); err != nil {
			continue
		}
		return append([]byte(nil), respuesta.Manifiesto...), nil
	}
	return nil, fmt.Errorf("RootCID no encontrado: %s", rootCID)
}

func (nodo *Nodo) recibirOrganismo(stream network.Stream) {
	defer stream.Close()

	var solicitud solicitudOrganismo
	if err := json.NewDecoder(bufio.NewReader(stream)).Decode(&solicitud); err != nil {
		return
	}
	rootCID := normalizarRootCID(solicitud.RootCID)

	if solicitud.Operacion == "replicar" {
		if err := nodo.AnunciarOrganismo(context.Background(), rootCID, solicitud.Manifiesto); err != nil {
			_ = json.NewEncoder(stream).Encode(respuestaOrganismo{RootCID: rootCID, Error: err.Error()})
			return
		}
		_ = json.NewEncoder(stream).Encode(respuestaOrganismo{RootCID: rootCID})
		return
	}

	manifiesto, err := nodo.almacen.Obtener(rootCID)
	if err != nil {
		_ = json.NewEncoder(stream).Encode(respuestaOrganismo{RootCID: rootCID, Error: "contenido no disponible o inválido"})
		return
	}

	_ = json.NewEncoder(stream).Encode(respuestaOrganismo{
		RootCID:    rootCID,
		Manifiesto: manifiesto,
	})
}

func (nodo *Nodo) RecuperarOrganismo(ctx context.Context, rootCID string) ([]byte, error) {
	rootCID = normalizarRootCID(rootCID)
	if nodo.almacen != nil {
		if contenido, err := nodo.almacen.Obtener(rootCID); err == nil { return contenido, nil }
	}
	return nodo.buscarProveedor(ctx, rootCID)
}

// DisponibleLocal comprueba existencia e integridad del bloque local.
// No afirma que existan peers remotos ni que el contenido esté disponible en la red.
func (nodo *Nodo) DisponibleLocal(rootCID string) (bool, error) {
	if nodo == nil || nodo.almacen == nil { return false, fmt.Errorf("almacén de contenido no inicializado") }
	return nodo.almacen.Disponible(normalizarRootCID(rootCID))
}
