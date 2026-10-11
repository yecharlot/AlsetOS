package organismo

import "fmt"

// Transiciones permitidas del ciclo de vida (Contrato / Manifiesto de organismos).
// Un organismo no puede pasar a Ejecutando si le faltan requisitos esenciales.

var transiciones = map[Estado][]Estado{
	Creado:     {Listo, Error, Detenido},
	Listo:      {Ejecutando, Detenido, Error},
	Ejecutando: {Detenido, Error, Listo},
	Detenido:   {Listo, Error},
	Error:      {Listo, Detenido},
}

// PuedeTransicionar indica si el salto de estado está permitido por el contrato.
func (o *Organismo) PuedeTransicionar(destino Estado) bool {
	if o == nil {
		return false
	}
	permitidos, ok := transiciones[o.Estado]
	if !ok {
		return false
	}
	for _, e := range permitidos {
		if e == destino {
			return true
		}
	}
	return false
}

// RequisitosParaEjecutar comprueba precondiciones del Contrato:
// identidad de definición (RootCID) y al menos una capacidad o memoria inicial explícita
// no son obligatorias en CREADO, pero EJECUTANDO exige RootCID no vacío
// salvo organismos de sistema marcados en memoria["sistema"]=="1".
func (o *Organismo) RequisitosParaEjecutar() error {
	if o == nil {
		return fmt.Errorf("organismo nulo")
	}
	if o.Memoria != nil && o.Memoria["sistema"] == "1" {
		return nil
	}
	if o.RootCID == "" {
		return fmt.Errorf("no se puede ejecutar sin RootCID (identidad de definición)")
	}
	return nil
}

// Transicionar aplica una transición de ciclo de vida.
// Si destino es Ejecutando, exige RequisitosParaEjecutar.
func (o *Organismo) Transicionar(destino Estado) error {
	if o == nil {
		return fmt.Errorf("organismo nulo")
	}
	if !o.PuedeTransicionar(destino) {
		return fmt.Errorf("transición ilegal: %s → %s", o.Estado, destino)
	}
	if destino == Ejecutando {
		if err := o.RequisitosParaEjecutar(); err != nil {
			return err
		}
	}
	o.Estado = destino
	return nil
}

// Preparar marca el organismo Listo tras asignar RootCID (definición canónica).
func (o *Organismo) Preparar(rootCID string) error {
	if o == nil {
		return fmt.Errorf("organismo nulo")
	}
	if rootCID == "" {
		return fmt.Errorf("RootCID vacío")
	}
	o.RootCID = rootCID
	if o.Estado == Creado {
		return o.Transicionar(Listo)
	}
	if o.Estado == Listo {
		return nil
	}
	return fmt.Errorf("Preparar solo desde creado o listo (estado=%s)", o.Estado)
}

// Iniciar pasa a Ejecutando si el contrato lo permite.
func (o *Organismo) Iniciar() error {
	return o.Transicionar(Ejecutando)
}

// Detener pasa a Detenido desde estados activos.
func (o *Organismo) Detener() error {
	if o == nil {
		return fmt.Errorf("organismo nulo")
	}
	if o.Estado != Ejecutando && o.Estado != Listo && o.Estado != Error {
		return fmt.Errorf("no se puede detener desde %s", o.Estado)
	}
	if o.Estado == Listo {
		return o.Transicionar(Detenido)
	}
	return o.Transicionar(Detenido)
}

// Fallar registra estado Error (recuperación posterior vía Listo).
func (o *Organismo) Fallar(motivo string) error {
	if o == nil {
		return fmt.Errorf("organismo nulo")
	}
	if o.Memoria == nil {
		o.Memoria = make(map[string]string)
	}
	if motivo != "" {
		o.Memoria["ultimo_error"] = motivo
	}
	if o.Estado == Error {
		return nil
	}
	if !o.PuedeTransicionar(Error) {
		return fmt.Errorf("no se puede fallar desde %s", o.Estado)
	}
	o.Estado = Error
	return nil
}

// OtorgarCapacidad declara una capacidad (autorización explícita, no decisión Mind).
func (o *Organismo) OtorgarCapacidad(id string) {
	if o == nil || id == "" {
		return
	}
	if o.Capacidad == nil {
		o.Capacidad = make(map[string]bool)
	}
	o.Capacidad[id] = true
}

// TieneCapacidad consulta autorización explícita.
func (o *Organismo) TieneCapacidad(id string) bool {
	if o == nil || o.Capacidad == nil {
		return false
	}
	return o.Capacidad[id]
}
