package identidad

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Identidad struct {
	ID string
	ClavePublica ed25519.PublicKey
	ClavePrivada ed25519.PrivateKey
}

type documento struct {
	Publica string `json:"publica"`
	Privada string `json:"privada"`
}

func Nueva() (*Identidad, error) {
	publica, privada, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generar identidad: %w", err)
	}
	return crear(publica, privada), nil
}

func Cargar(ruta string) (*Identidad, error) {
	contenido, err := os.ReadFile(ruta)
	if err != nil {
		return nil, fmt.Errorf("leer identidad: %w", err)
	}
	var doc documento
	if err := json.Unmarshal(contenido, &doc); err != nil {
		return nil, fmt.Errorf("analizar identidad: %w", err)
	}
	publica, err := base64.StdEncoding.DecodeString(doc.Publica)
	if err != nil {
		return nil, fmt.Errorf("decodificar clave pública: %w", err)
	}
	privada, err := base64.StdEncoding.DecodeString(doc.Privada)
	if err != nil {
		return nil, fmt.Errorf("decodificar clave privada: %w", err)
	}
	if len(publica) != ed25519.PublicKeySize || len(privada) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("tamaño de clave de identidad inválido")
	}
	return crear(ed25519.PublicKey(publica), ed25519.PrivateKey(privada)), nil
}

func (identidad *Identidad) Guardar(ruta string) error {
	if identidad == nil || len(identidad.ClavePublica) != ed25519.PublicKeySize || len(identidad.ClavePrivada) != ed25519.PrivateKeySize {
		return fmt.Errorf("identidad inválida")
	}
	doc := documento{
		Publica: base64.StdEncoding.EncodeToString(identidad.ClavePublica),
		Privada: base64.StdEncoding.EncodeToString(identidad.ClavePrivada),
	}
	contenido, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar identidad: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(ruta), 0700); err != nil {
		return fmt.Errorf("crear directorio de identidad: %w", err)
	}
	temporal := ruta + ".tmp"
	if err := os.WriteFile(temporal, contenido, 0600); err != nil {
		return fmt.Errorf("guardar identidad temporal: %w", err)
	}
	if err := os.Rename(temporal, ruta); err != nil {
		return fmt.Errorf("confirmar identidad: %w", err)
	}
	return nil
}

func (identidad *Identidad) Firmar(contenido []byte) []byte {
	return ed25519.Sign(identidad.ClavePrivada, contenido)
}

func Verificar(clavePublica, contenido, firma []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(clavePublica), contenido, firma)
}

func crear(publica ed25519.PublicKey, privada ed25519.PrivateKey) *Identidad {
	idBytes := publica
	return &Identidad{
		ID: "node:" + base64.RawURLEncoding.EncodeToString(idBytes),
		ClavePublica: publica,
		ClavePrivada: privada,
	}
}
