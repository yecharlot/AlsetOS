package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"

	"github.com/yecharlot/AlsetOS/rootcid"
)

const ProtocoloOrganismo = "/alset/organism/1.0.0"

type solicitudOrganismo struct {
	RootCID string `json:"rootcid"`
}

type respuestaOrganismo struct {
	RootCID    string `json:"rootcid"`
	Manifiesto []byte `json:"manifiesto"`
}

func normalizarRootCID(valor string) string {
	return strings.TrimSpace(valor)
}

func (nodo *Nodo) AnunciarOrganismo(ctx context.Context, rootCID string, manifiesto []byte) error {
	rootCID = normalizarRootCID(rootCID)
	if rootCID == "" {
		return fmt.Errorf("RootCID vacío")
	}
	if rootcid.CrearContenido(manifiesto) != rootCID {
		return fmt.Errorf("el manifiesto no corresponde al RootCID")
	}

	nodo.mu.Lock()
	nodo.organismos[rootCID] = append([]byte(nil), manifiesto...)
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
			nodo.mu.RLock()
			local := append([]byte(nil), nodo.organismos[rootCID]...)
			nodo.mu.RUnlock()
			if len(local) > 0 {
				return local, nil
			}
		}
		if err := nodo.Host.Connect(ctx, proveedor); err != nil {
			continue
		}
		stream, err := nodo.Host.NewStream(ctx, proveedor.ID, ProtocoloOrganismo)
		if err != nil {
			continue
		}
		_ = json.NewEncoder(stream).Encode(solicitudOrganismo{RootCID: rootCID})
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
		return respuesta.Manifiesto, nil
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

	nodo.mu.RLock()
	manifiesto := append([]byte(nil), nodo.organismos[rootCID]...)
	nodo.mu.RUnlock()
	if len(manifiesto) == 0 {
		return
	}

	_ = json.NewEncoder(stream).Encode(respuestaOrganismo{
		RootCID: rootCID,
		Manifiesto: manifiesto,
	})
}

func (nodo *Nodo) RecuperarOrganismo(ctx context.Context, rootCID string) ([]byte, error) {
	if manifiesto, err := nodo.buscarProveedor(ctx, normalizarRootCID(rootCID)); err == nil {
		return manifiesto, nil
	} else {
		return nil, err
	}
}

func copiarOrganismo(r io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, 4<<20))
}
