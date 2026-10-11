package organismo

// Estado representa el ciclo de vida observable de un organismo (contrato único Go + kernel_native).
// Valores canónicos (español, minúsculas). El campo phase del kernel C DEBE usar los mismos literales.
type Estado string

const (
	Creado     Estado = "creado"
	Listo      Estado = "listo"
	Ejecutando Estado = "ejecutando"
	Detenido   Estado = "detenido"
	Error      Estado = "error"
)

// Espacio de ejecución: frontera kernel vs usuario.
type Espacio string

const (
	EspacioKernel Espacio = "kernel" // fundacional / runtime del nodo
	EspacioUsuario Espacio = "usuario" // apps / ORGES de usuario
)

// Organismo representa una unidad ejecutable persistente de AlsetOS.
type Organismo struct {
	Nombre string
	RootCID string
	// DefinicionCID es el CID de la definición canónica verificada (puede coincidir con RootCID).
	// RootCID no vacío NO implica que DefinicionVerificada sea true.
	DefinicionCID string
	DefinicionVerificada bool
	Memoria map[string]string
	Capacidad map[string]bool
	Estado Estado
	// Fundacional: sellado en creación (NuevoFundacional). NO se deriva de Memoria mutable.
	Fundacional bool
	Espacio Espacio
	// Autoridad que otorgó el sello fundacional (oid/nombre), vacío si no aplica.
	FundacionalPor string
}

// Nuevo crea un organismo de usuario (espacio usuario, no fundacional).
func Nuevo(nombre string) *Organismo {
	return &Organismo{
		Nombre:    nombre,
		Memoria:   make(map[string]string),
		Capacidad: make(map[string]bool),
		Estado:    Creado,
		Espacio:   EspacioUsuario,
	}
}

// NuevoFundacional crea un organismo del runtime del nodo.
// Solo el código del núcleo debe llamar esto; no hay bypass por propiedades de Memoria.
func NuevoFundacional(nombre, selladoPor string) *Organismo {
	o := Nuevo(nombre)
	o.Fundacional = true
	o.Espacio = EspacioKernel
	o.FundacionalPor = selladoPor
	return o
}
