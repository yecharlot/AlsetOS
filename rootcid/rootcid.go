package rootcid

import (
	"crypto/sha256"
	"encoding/hex"
)

// Crear genera una identidad determinista para un organismo por nombre.
// Se conserva como API de compatibilidad para el núcleo inicial.
func Crear(nombre string) string {
	return CrearContenido([]byte("organismo:" + nombre + ":alsetos:v0.1"))
}

// CrearContenido genera una identidad por contenido.
// El mismo contenido produce siempre el mismo RootCID.
func CrearContenido(contenido []byte) string {
	resumen := sha256.Sum256(contenido)
	return "rootcid:" + hex.EncodeToString(resumen[:])
}
