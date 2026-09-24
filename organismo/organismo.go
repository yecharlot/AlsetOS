package organismo

// Organismo representa una unidad ejecutable persistente de AlsetOS.
type Organismo struct {
	Nombre    string
	RootCID   string
	Memoria   map[string]string
	Capacidad map[string]bool
}

// Nuevo crea un organismo con memoria y capacidades iniciales vacías.
func Nuevo(nombre string) *Organismo {
	return &Organismo{
		Nombre:    nombre,
		Memoria:   make(map[string]string),
		Capacidad: make(map[string]bool),
	}
}
