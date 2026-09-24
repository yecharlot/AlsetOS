package organismo

// Estado representa el ciclo de vida observable de un organismo.
type Estado string

const (
	Creado Estado = "creado"
	Listo Estado = "listo"
	Ejecutando Estado = "ejecutando"
	Detenido Estado = "detenido"
	Error Estado = "error"
)

// Organismo representa una unidad ejecutable persistente de AlsetOS.
type Organismo struct {
	Nombre string
	RootCID string
	Memoria map[string]string
	Capacidad map[string]bool
	Estado Estado
}

// Nuevo crea un organismo con memoria y capacidades iniciales vacías.
func Nuevo(nombre string) *Organismo {
	return &Organismo{
		Nombre: nombre,
		Memoria: make(map[string]string),
		Capacidad: make(map[string]bool),
		Estado: Creado,
	}
}
