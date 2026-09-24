package rootcid

import (
	"crypto/sha256"
	"encoding/hex"
)

// Crear genera una identidad determinista para un organismo.
func Crear(nombre string) string {
	resumen := sha256.Sum256([]byte("organismo:" + nombre + ":alsetos:v0.1"))
	return "rootcid:" + hex.EncodeToString(resumen[:])
}
