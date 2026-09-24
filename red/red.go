package red

import (
	"bufio"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/pulso"
)

type Conexion struct {
	conexion io.ReadWriteCloser
	lector *bufio.Reader
}

type Mensaje struct {
	NodoID string
	Publica string
	Pulso pulso.Pulso
	Firma string
}

type mensajeFirmable struct {
	NodoID string
	Publica string
	Pulso pulso.Pulso
}

func NuevaConexion(conexion io.ReadWriteCloser) *Conexion {
	return &Conexion{conexion: conexion, lector: bufio.NewReader(conexion)}
}

func (conexion *Conexion) EnviarPulso(evento pulso.Pulso) error {
	contenido, err := json.Marshal(evento)
	if err != nil { return fmt.Errorf("serializar pulso: %w", err) }
	contenido = append(contenido, '\n')
	if _, err := conexion.conexion.Write(contenido); err != nil { return fmt.Errorf("enviar pulso: %w", err) }
	return nil
}

func (conexion *Conexion) EnviarPulsoFirmado(identidadNodo *identidad.Identidad, evento pulso.Pulso) error {
	if identidadNodo == nil { return fmt.Errorf("identidad nula") }
	mensaje := CrearMensajeFirmado(identidadNodo, evento)
	contenido, err := json.Marshal(mensaje)
	if err != nil { return fmt.Errorf("serializar mensaje firmado: %w", err) }
	contenido = append(contenido, '\n')
	if _, err := conexion.conexion.Write(contenido); err != nil { return fmt.Errorf("enviar Pulse firmado: %w", err) }
	return nil
}

func CrearMensajeFirmado(identidadNodo *identidad.Identidad, evento pulso.Pulso) Mensaje {
	firmable := mensajeFirmable{
		NodoID: identidadNodo.ID,
		Publica: base64.StdEncoding.EncodeToString(identidadNodo.ClavePublica),
		Pulso: evento,
	}
	contenido, _ := json.Marshal(firmable)
	return Mensaje{
		NodoID: firmable.NodoID,
		Publica: firmable.Publica,
		Pulso: firmable.Pulso,
		Firma: base64.StdEncoding.EncodeToString(identidadNodo.Firmar(contenido)),
	}
}

func VerificarMensaje(mensaje Mensaje) error {
	publica, err := base64.StdEncoding.DecodeString(mensaje.Publica)
	if err != nil { return fmt.Errorf("decodificar clave pública: %w", err) }
	firma, err := base64.StdEncoding.DecodeString(mensaje.Firma)
	if err != nil { return fmt.Errorf("decodificar firma: %w", err) }
	if len(publica) != ed25519.PublicKeySize || len(firma) != ed25519.SignatureSize {
		return fmt.Errorf("credencial criptográfica inválida")
	}
	firmable := mensajeFirmable{NodoID: mensaje.NodoID, Publica: mensaje.Publica, Pulso: mensaje.Pulso}
	contenido, err := json.Marshal(firmable)
	if err != nil { return fmt.Errorf("serializar verificación: %w", err) }
	if !identidad.Verificar(publica, contenido, firma) { return fmt.Errorf("firma de Pulse inválida") }
	return nil
}

func (conexion *Conexion) RecibirPulso() (pulso.Pulso, error) {
	contenido, err := conexion.lector.ReadBytes('\n')
	if err != nil { return pulso.Pulso{}, fmt.Errorf("recibir pulso: %w", err) }
	var evento pulso.Pulso
	if err := json.Unmarshal(contenido, &evento); err != nil { return pulso.Pulso{}, fmt.Errorf("analizar pulso: %w", err) }
	return evento, nil
}

func (conexion *Conexion) RecibirPulsoFirmado() (Mensaje, error) {
	contenido, err := conexion.lector.ReadBytes('\n')
	if err != nil { return Mensaje{}, fmt.Errorf("recibir Pulse firmado: %w", err) }
	var mensaje Mensaje
	if err := json.Unmarshal(contenido, &mensaje); err != nil { return Mensaje{}, fmt.Errorf("analizar Pulse firmado: %w", err) }
	if err := VerificarMensaje(mensaje); err != nil { return Mensaje{}, err }
	return mensaje, nil
}

func (conexion *Conexion) Cerrar() error {
	return conexion.conexion.Close()
}
