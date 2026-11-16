//go:build linux

package p2p

import (
	"fmt"
	"syscall"
)

func medirSistemaArchivos(ruta string) (CapacidadAlmacenamiento, error) {
	var st syscall.Statfs_t
	if ruta == "" { return CapacidadAlmacenamiento{}, fmt.Errorf("el almacén está en memoria; no existe capacidad física que medir") }
	if err := syscall.Statfs(ruta, &st); err != nil { return CapacidadAlmacenamiento{}, fmt.Errorf("medir capacidad del almacén: %w", err) }
	if st.Bsize <= 0 { return CapacidadAlmacenamiento{}, fmt.Errorf("tamaño de bloque del sistema de archivos inválido") }
	total := uint64(st.Blocks) * uint64(st.Bsize)
	libre := uint64(st.Bavail) * uint64(st.Bsize)
	if total == 0 || libre > total { return CapacidadAlmacenamiento{}, fmt.Errorf("estadísticas del sistema de archivos inconsistentes") }
	return CapacidadAlmacenamiento{Ruta:ruta, TotalBytes:int64(total), DisponibleBytes:int64(libre), ReservableBytes:int64(libre/5*4 + (libre%5)*4/5)}, nil
}
