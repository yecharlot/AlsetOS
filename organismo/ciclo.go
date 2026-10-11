package organismo

import "fmt"

// Contrato de ciclo de vida (ATS / Manifiesto).
//
// Significado de Ejecutando:
//   - Go runtime: el organismo está bajo el motor/frontera y puede recibir ticks/Pulse.
//   - kernel_native: el OCB está activo en el scheduler del nodo (misma palabra "ejecutando").
//
// Precondiciones a Ejecutando:
//   - Transición legal según mapa.
//   - Fundacional: puede ejecutar sin RootCID de app (identidad de nodo sellada).
//   - Usuario: exige RootCID presente Y DefinicionVerificada (no basta cadena no vacía).

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

// RequisitosParaEjecutar comprueba precondiciones del Contrato.
func (o *Organismo) RequisitosParaEjecutar() error {
	if o == nil {
		return fmt.Errorf("organismo nulo")
	}
	// Fundacional: política explícita sellada en creación (no Memoria["sistema"]).
	if o.Fundacional && o.Espacio == EspacioKernel {
		return nil
	}
	if o.RootCID == "" {
		return fmt.Errorf("no se puede ejecutar sin RootCID (presencia de identidad)")
	}
	if !o.DefinicionVerificada {
		return fmt.Errorf("RootCID presente pero definición canónica no verificada")
	}
	return nil
}

// Transicionar aplica una transición de ciclo de vida.
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

// Preparar asigna RootCID observado (presencia). NO marca verificación canónica.
func (o *Organismo) Preparar(rootCID string) error {
	if o == nil {
		return fmt.Errorf("organismo nulo")
	}
	if rootCID == "" {
		return fmt.Errorf("RootCID vacío")
	}
	o.RootCID = rootCID
	o.DefinicionVerificada = false
	if o.Estado == Creado {
		return o.Transicionar(Listo)
	}
	if o.Estado == Listo {
		return nil
	}
	return fmt.Errorf("Preparar solo desde creado o listo (estado=%s)", o.Estado)
}

// VerificarDefinicion marca la definición canónica solo si el CID coincide con el contenido.
// Llamar con el resultado de rootcid.Verificar(contenido, cid).
func (o *Organismo) VerificarDefinicion(definicionCID string, ok bool) error {
	if o == nil {
		return fmt.Errorf("organismo nulo")
	}
	if !ok || definicionCID == "" {
		o.DefinicionVerificada = false
		return fmt.Errorf("definición canónica no válida")
	}
	o.DefinicionCID = definicionCID
	if o.RootCID == "" {
		o.RootCID = definicionCID
	}
	if o.RootCID != definicionCID {
		o.DefinicionVerificada = false
		return fmt.Errorf("RootCID (%s) ≠ definición verificada (%s)", o.RootCID, definicionCID)
	}
	o.DefinicionVerificada = true
	if o.Estado == Creado {
		return o.Transicionar(Listo)
	}
	return nil
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

// OtorgarCapacidad es de bajo nivel; preferir capacidad.Otorgar(autoridad, objetivo, id).
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
