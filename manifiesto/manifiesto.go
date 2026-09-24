package manifiesto

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/yecharlot/AlsetOS/zyrion"
)

type ConfiguracionZyrion struct { Estado string; Estrategia string }
type Manifiesto struct { Nombre string; LispAI []string; Zyrion ConfiguracionZyrion; Capacidades []string; Genes []string; Memoria string; Agentes []string }

func Cargar(ruta string) (Manifiesto, error) {
	contenido, err := os.ReadFile(ruta)
	if err != nil { return Manifiesto{}, fmt.Errorf("leer manifiesto: %w", err) }
	var documento Manifiesto
	if err := json.Unmarshal(contenido, &documento); err != nil { return Manifiesto{}, fmt.Errorf("analizar manifiesto: %w", err) }
	if documento.Nombre == "" { return Manifiesto{}, fmt.Errorf("el manifiesto requiere nombre") }
	return documento, nil
}

func (documento Manifiesto) EstadoZyrion() (zyrion.Valor, error) {
	switch documento.Zyrion.Estado {
	case "no": return zyrion.No, nil
	case "si": return zyrion.Si, nil
	case "incierto": return zyrion.Incierto, nil
	default: return zyrion.Incierto, fmt.Errorf("estado Zyrion desconocido: %q", documento.Zyrion.Estado)
	}
}
