package p2p

import "fmt"

// CapacidadAlmacenamiento describe una medición puntual del sistema de archivos.
// ReservableBytes conserva un 20% del espacio libre como margen operativo.
type CapacidadAlmacenamiento struct {
	Ruta string `json:"ruta"`
	TotalBytes int64 `json:"total_bytes"`
	DisponibleBytes int64 `json:"disponible_bytes"`
	ReservableBytes int64 `json:"reservable_bytes"`
}

// MedirCapacidadRetencion no cambia políticas ni reserva espacio. Es una
// observación puntual del sistema de archivos que contiene los bloques.
func (n *Nodo) MedirCapacidadRetencion() (CapacidadAlmacenamiento, error) {
	if n == nil || n.almacen == nil { return CapacidadAlmacenamiento{}, fmt.Errorf("almacén no inicializado") }
	return medirSistemaArchivos(n.almacen.dir)
}

// ConfigurarCapacidadRetencionDisponible aplica explícitamente el 80% del
// espacio libre medido como techo lógico de compromisos. No reserva bytes.
func (n *Nodo) ConfigurarCapacidadRetencionDisponible() (CapacidadAlmacenamiento, error) {
	medicion, err := n.MedirCapacidadRetencion()
	if err != nil { return CapacidadAlmacenamiento{}, err }
	if medicion.ReservableBytes <= 0 { return medicion, fmt.Errorf("no hay capacidad conservadora disponible para compromisos") }
	if err := n.ConfigurarCapacidadRetencion(medicion.ReservableBytes); err != nil { return medicion, err }
	return medicion, nil
}
