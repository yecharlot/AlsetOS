package p2p

import (
	"encoding/json"
	"fmt"
	"os"
)

func cargarOrganismos(ruta string) (map[string][]byte, error) {
	resultado := make(map[string][]byte)
	if ruta == "" {
		return resultado, nil
	}

	contenido, err := os.ReadFile(ruta)
	if os.IsNotExist(err) {
		return resultado, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer almacén de organismos: %w", err)
	}
	if len(contenido) == 0 {
		return resultado, nil
	}
	if err := json.Unmarshal(contenido, &resultado); err != nil {
		return nil, fmt.Errorf("analizar almacén de organismos: %w", err)
	}
	if resultado == nil {
		resultado = make(map[string][]byte)
	}
	return resultado, nil
}

func guardarOrganismos(ruta string, organismos map[string][]byte) error {
	if ruta == "" {
		return nil
	}
	contenido, err := json.MarshalIndent(organismos, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar almacén de organismos: %w", err)
	}
	if err := os.MkdirAll(dirRuta(ruta), 0700); err != nil {
		return fmt.Errorf("crear directorio del almacén: %w", err)
	}

	temporal := ruta + ".tmp"
	if err := os.WriteFile(temporal, contenido, 0600); err != nil {
		return fmt.Errorf("escribir almacén temporal: %w", err)
	}
	if err := os.Rename(temporal, ruta); err != nil {
		return fmt.Errorf("confirmar almacén de organismos: %w", err)
	}
	return nil
}

func dirRuta(ruta string) string {
	for i := len(ruta) - 1; i >= 0; i-- {
		if ruta[i] == '/' {
			if i == 0 { return "/" }
			return ruta[:i]
		}
	}
	return "."
}
