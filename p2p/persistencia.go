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

	temporal, err := os.CreateTemp(dirRuta(ruta), ".organismos-*")
	if err != nil { return fmt.Errorf("crear almacén temporal: %w", err) }
	nombreTemporal := temporal.Name()
	confirmado := false
	defer func() { if !confirmado { _ = os.Remove(nombreTemporal) } }()
	if err := temporal.Chmod(0600); err != nil { _ = temporal.Close(); return err }
	if _, err := temporal.Write(contenido); err != nil { _ = temporal.Close(); return fmt.Errorf("escribir almacén temporal: %w", err) }
	if err := temporal.Sync(); err != nil { _ = temporal.Close(); return fmt.Errorf("sincronizar almacén temporal: %w", err) }
	if err := temporal.Close(); err != nil { return fmt.Errorf("cerrar almacén temporal: %w", err) }
	if err := os.Rename(nombreTemporal, ruta); err != nil { return fmt.Errorf("confirmar almacén de organismos: %w", err) }
	if directorio, err := os.Open(dirRuta(ruta)); err == nil { _ = directorio.Sync(); _ = directorio.Close() }
	confirmado = true
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
