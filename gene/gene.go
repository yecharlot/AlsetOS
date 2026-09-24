package gene

import "github.com/yecharlot/AlsetOS/organismo"

// Gene representa una capacidad ejecutable.
type Gene struct {
	Nombre string
}

// Ejecutar ejecuta la capacidad y registra el último resultado en memoria.
func (gen Gene) Ejecutar(entidad *organismo.Organismo) string {
	resultado := "Gene " + gen.Nombre + " ejecutado para " + entidad.Nombre
	entidad.Memoria["ultimo_resultado"] = resultado
	return resultado
}

// Registro contiene Genes disponibles para un organismo.
type Registro struct {
	genes map[string]Gene
}

// NuevoRegistro crea un registro vacío de Genes.
func NuevoRegistro() *Registro {
	return &Registro{genes: make(map[string]Gene)}
}

// Registrar incorpora un Gene por nombre.
func (registro *Registro) Registrar(gen Gene) {
	registro.genes[gen.Nombre] = gen
}

// Obtener recupera un Gene por nombre.
func (registro *Registro) Obtener(nombre string) (Gene, bool) {
	gen, existe := registro.genes[nombre]
	return gen, existe
}
