package rootcid

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Crear genera una identidad determinista para un organismo por nombre.
func Crear(nombre string) string {
	return CrearContenido([]byte("organismo:" + nombre + ":alsetos:v0.1"))
}

// CrearContenido genera una identidad por contenido.
func CrearContenido(contenido []byte) string {
	resumen := sha256.Sum256(contenido)
	return "rootcid:" + hex.EncodeToString(resumen[:])
}

// Presente solo indica que la cadena de identidad no está vacía.
// NO implica verificación de definición canónica.
func Presente(cid string) bool {
	return strings.TrimSpace(cid) != ""
}

// Verificar comprueba que cid == CrearContenido(contenido).
// Es la distinción ATS: presencia ≠ verificación.
func Verificar(contenido []byte, cid string) error {
	if !Presente(cid) {
		return fmt.Errorf("RootCID ausente")
	}
	esperado := CrearContenido(contenido)
	if cid != esperado {
		return fmt.Errorf("RootCID no coincide con definición canónica")
	}
	return nil
}

// VerificarOK es un helper booleano para organismo.VerificarDefinicion.
func VerificarOK(contenido []byte, cid string) bool {
	return Verificar(contenido, cid) == nil
}
