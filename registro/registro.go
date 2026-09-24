package registro

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Entrada struct {
	Nombre string `json:"nombre"`
	RootCID string `json:"rootcid"`
	Estado string `json:"estado"`
	Actualizado string `json:"actualizado"`
}

type Registro struct {
	mu       sync.RWMutex
	ruta     string
	entradas map[string]Entrada
}

func Nuevo(ruta string) (*Registro, error) {
	registro := &Registro{
		ruta:     ruta,
		entradas: make(map[string]Entrada),
	}

	contenido, err := os.ReadFile(ruta)
	if os.IsNotExist(err) {
		return registro, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer registro: %w", err)
	}
	if len(contenido) == 0 {
		return registro, nil
	}
	if err := json.Unmarshal(contenido, &registro.entradas); err != nil {
		return nil, fmt.Errorf("analizar registro: %w", err)
	}
	if registro.entradas == nil {
		registro.entradas = make(map[string]Entrada)
	}

	return registro, nil
}

func (registro *Registro) Guardar(entrada Entrada) error {
	registro.mu.Lock()
	defer registro.mu.Unlock()

	registro.entradas[entrada.RootCID] = entrada
	contenido, err := json.MarshalIndent(registro.entradas, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar registro: %w", err)
	}

	if err := os.MkdirAll(dir(registro.ruta), 0755); err != nil {
		return fmt.Errorf("crear directorio del registro: %w", err)
	}

	temporal := registro.ruta + ".tmp"
	if err := os.WriteFile(temporal, contenido, 0600); err != nil {
		return fmt.Errorf("escribir registro temporal: %w", err)
	}
	if err := os.Rename(temporal, registro.ruta); err != nil {
		return fmt.Errorf("confirmar registro: %w", err)
	}

	return nil
}

func (registro *Registro) Obtener(rootCID string) (Entrada, bool) {
	registro.mu.RLock()
	defer registro.mu.RUnlock()

	entrada, existe := registro.entradas[rootCID]
	return entrada, existe
}

func (registro *Registro) Todos() []Entrada {
	registro.mu.RLock()
	defer registro.mu.RUnlock()

	salida := make([]Entrada, 0, len(registro.entradas))
	for _, entrada := range registro.entradas {
		salida = append(salida, entrada)
	}
	return salida
}

func dir(ruta string) string {
	for i := len(ruta) - 1; i >= 0; i-- {
		if ruta[i] == '/' {
			if i == 0 {
				return "/"
			}
			return ruta[:i]
		}
	}
	return "."
}
