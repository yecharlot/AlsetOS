package memoria

import (
	"encoding/json"
	"fmt"
	"os"
)

// Guardar persiste una memoria de organismo en JSON.
func Guardar(ruta string, datos map[string]string) error {
	contenido, err := json.MarshalIndent(datos, "", "  ")
	if err != nil { return fmt.Errorf("serializar memoria: %w", err) }
	if err := os.WriteFile(ruta, contenido, 0600); err != nil { return fmt.Errorf("guardar memoria: %w", err) }
	return nil
}

// Cargar recupera una memoria persistida. Si el archivo no existe devuelve memoria vacía.
func Cargar(ruta string) (map[string]string, error) {
	contenido, err := os.ReadFile(ruta)
	if os.IsNotExist(err) { return make(map[string]string), nil }
	if err != nil { return nil, fmt.Errorf("leer memoria: %w", err) }
	var datos map[string]string
	if err := json.Unmarshal(contenido, &datos); err != nil { return nil, fmt.Errorf("analizar memoria: %w", err) }
	if datos == nil { datos = make(map[string]string) }
	return datos, nil
}
