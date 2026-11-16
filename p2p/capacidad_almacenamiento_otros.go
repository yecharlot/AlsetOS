//go:build !linux

package p2p

import "fmt"

func medirSistemaArchivos(ruta string) (CapacidadAlmacenamiento, error) {
	return CapacidadAlmacenamiento{}, fmt.Errorf("medición física de almacenamiento no implementada para esta plataforma")
}
