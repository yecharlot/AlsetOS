package capacidad

import "github.com/yecharlot/AlsetOS/organismo"

// Tiene verifica una capacidad explícita del organismo.
func Tiene(entidad *organismo.Organismo, nombre string) bool {
	return entidad.Capacidad[nombre]
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
