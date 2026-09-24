package manifiesto

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yecharlot/AlsetOS/zyrion"
)

type ConfiguracionZyrion struct {
	Estado string `json:"estado"`
	Estrategia string `json:"estrategia,omitempty"`
}

type Manifiesto struct {
	Nombre string `json:"nombre"`
	LispAI []string `json:"lispai,omitempty"`
	Zyrion ConfiguracionZyrion `json:"zyrion"`
	Capacidades []string `json:"capacidades,omitempty"`
	Genes []string `json:"genes,omitempty"`
	Memoria string `json:"memoria,omitempty"`
	Agentes []string `json:"agentes,omitempty"`
}

func Cargar(ruta string) (Manifiesto, error) {
	contenido, err := os.ReadFile(ruta)
	if err != nil {
		return Manifiesto{}, fmt.Errorf("leer manifiesto: %w", err)
	}

	var documento Manifiesto
	if err := json.Unmarshal(contenido, &documento); err != nil {
		return Manifiesto{}, fmt.Errorf("analizar manifiesto: %w", err)
	}
	if err := documento.Validar(); err != nil {
		return Manifiesto{}, err
	}
	return documento, nil
}

func (documento Manifiesto) Validar() error {
	if documento.Nombre == "" {
		return fmt.Errorf("el manifiesto requiere nombre")
	}
	if documento.Zyrion.Estado == "" {
		return fmt.Errorf("el manifiesto requiere estado Zyrion")
	}
	if _, err := documento.EstadoZyrion(); err != nil {
		return err
	}
	return nil
}

func (documento Manifiesto) EstadoZyrion() (zyrion.Valor, error) {
	switch documento.Zyrion.Estado {
	case "no":
		return zyrion.No, nil
	case "si":
		return zyrion.Si, nil
	case "incierto":
		return zyrion.Incierto, nil
	default:
		return zyrion.Incierto, fmt.Errorf("estado Zyrion desconocido: %q", documento.Zyrion.Estado)
	}
}

// Canonico serializa la definición semántica del organismo de forma estable.
// El resultado es la base del RootCID: cambiar la definición cambia la identidad.
func (documento Manifiesto) Canonico() ([]byte, error) {
	if err := documento.Validar(); err != nil {
		return nil, err
	}
	contenido, err := json.Marshal(documento)
	if err != nil {
		return nil, fmt.Errorf("serializar manifiesto canónico: %w", err)
	}
	return contenido, nil
}
