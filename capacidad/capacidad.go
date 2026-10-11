package capacidad

import (
	"fmt"

	"github.com/yecharlot/AlsetOS/organismo"
)

// Capacidad de otorgar a otros (autoridad).
const CapOtorgar = "caps.otorgar"

// Tiene verifica una capacidad explícita del organismo.
func Tiene(entidad *organismo.Organismo, nombre string) bool {
	if entidad == nil {
		return false
	}
	if entidad.Fundacional && entidad.Espacio == organismo.EspacioKernel {
		// Fundacionales del nodo pueden inspeccionar; la ejecución de genes sigue requiriendo cap explícita
		// salvo que se haya otorgado. Aquí solo lectura de mapa.
	}
	return entidad.Capacidad != nil && entidad.Capacidad[nombre]
}

// Requerir impide una operación cuando el organismo no posee la capacidad.
func Requerir(entidad *organismo.Organismo, nombre string) error {
	if !Tiene(entidad, nombre) {
		return &Falta{Nombre: nombre}
	}
	return nil
}

// Falta representa una capacidad no concedida.
type Falta struct{ Nombre string }

func (falta *Falta) Error() string {
	return "capacidad no concedida: " + falta.Nombre
}

// PuedeOtorgar: autoridad = fundacional kernel O posee caps.otorgar.
func PuedeOtorgar(autoridad *organismo.Organismo) bool {
	if autoridad == nil {
		return false
	}
	if autoridad.Fundacional && autoridad.Espacio == organismo.EspacioKernel {
		return true
	}
	return Tiene(autoridad, CapOtorgar)
}

// Otorgar: Mind puede proponer; solo Capacidad (esta función) autoriza.
// Gene/WASM no deben llamar esto directamente sin pasar por aquí.
func Otorgar(autoridad, objetivo *organismo.Organismo, nombre string) error {
	if !PuedeOtorgar(autoridad) {
		return fmt.Errorf("autoridad sin permiso para otorgar capacidades")
	}
	if objetivo == nil || nombre == "" {
		return fmt.Errorf("objetivo o capacidad inválidos")
	}
	objetivo.OtorgarCapacidad(nombre)
	return nil
}
