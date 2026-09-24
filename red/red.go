package red

import (
	"bufio"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/pulso"
)

type Conexion struct {
	conexion net.Conn
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

func NuevaConexion(conexion net.Conn) *Conexion {
	return &Conexion{conexion: conexion, lector: bufio.NewReader(conexion)}
}

func (conexion *Conexion) EnviarPulso(evento pulso.Pulso) error {
	contenido, err := json.Marshal(evento)
	if err != nil {
		return fmt.Errorf("serializar pulso: %w", err)
	}
	contenido = append(contenido, '\n')
	if _, err := conexion.conexion.Write(contenido); err != nil {
		return fmt.Errorf("enviar pulso: %w", err)
	}
	return nil
}

func (conexion *Conexion) EnviarPulsoFirmado(identidadNodo *identidad.Identidad, evento pulso.Pulso) error {
	if identidadNodo == nil {
		return fmt.Errorf("identidad nula")
	}
	firmable := mensajeFirmable{
		NodoID: identidadNodo.ID,
		Publica: base64.StdEncoding.EncodeToString(identidadNodo.ClavePublica),
		Pulso: evento,
	}
	contenido, err := json.Marshal(firmable)
	if err != nil {
		return fmt.Errorf("serializar mensaje firmable: %w", err)
	}
	mensaje := Mensaje{
		NodoID: firmable.NodoID,
		Publica: firmable.Publica,
		Pulso: firmable.Pulso,
		Firma: base64.StdEncoding.EncodeToString(identidadNodo.Firmar(contenido)),
	}
	contenido, err = json.Marshal(mensaje)
	if err != nil {
		return fmt.Errorf("serializar mensaje firmado: %w", err)
	}
	contenido = append(contenido, '\n')
	if _, err := conexion.conexion.Write(contenido); err != nil {
		return fmt.Errorf("enviar Pulse firmado: %w", err)
	}
	return nil
}

func (conexion *Conexion) RecibirPulso() (pulso.Pulso, error) {
	contenido, err := conexion.lector.ReadBytes('\n')
	if err != nil {
		return pulso.Pulso{}, fmt.Errorf("recibir pulso: %w", err)
	}
	var evento pulso.Pulso
	if err := json.Unmarshal(contenido, &evento); err != nil {
		return pulso.Pulso{}, fmt.Errorf("analizar pulso: %w", err)
	}
	return evento, nil
}

func (conexion *Conexion) RecibirPulsoFirmado() (Mensaje, error) {
	contenido, err := conexion.lector.ReadBytes('\n')
	if err != nil {
		return Mensaje{}, fmt.Errorf("recibir Pulse firmado: %w", err)
	}
	var mensaje Mensaje
	if err := json.Unmarshal(contenido, &mensaje); err != nil {
		return Mensaje{}, fmt.Errorf("analizar Pulse firmado: %w", err)
	}

	publica, err := base64.StdEncoding.DecodeString(mensaje.Publica)
	if err != nil {
		return Mensaje{}, fmt.Errorf("decodificar clave pública: %w", err)
	}
	firma, err := base64.StdEncoding.DecodeString(mensaje.Firma)
	if err != nil {
		return Mensaje{}, fmt.Errorf("decodificar firma: %w", err)
	}
	if len(publica) != ed25519.PublicKeySize || len(firma) != ed25519.SignatureSize {
		return Mensaje{}, fmt.Errorf("credencial criptográfica inválida")
	}

	firmable := mensajeFirmable{
		NodoID: mensaje.NodoID,
		Publica: mensaje.Publica,
		Pulso: mensaje.Pulso,
	}
	firmableContenido, err := json.Marshal(firmable)
	if err != nil {
		return Mensaje{}, fmt.Errorf("serializar verificación: %w", err)
	}
	if !identidad.Verificar(publica, firmableContenido, firma) {
		return Mensaje{}, fmt.Errorf("firma de Pulse inválida")
	}
	return mensaje, nil
}

func (conexion *Conexion) Cerrar() error {
	return conexion.conexion.Close()
}
