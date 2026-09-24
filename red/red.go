package red

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/yecharlot/AlsetOS/pulso"
)

// Conexion representa el transporte mínimo de Pulse entre nodos.
// La semántica de AlsetOS no depende de TCP; esta capa puede sustituirse
// posteriormente por libp2p sin cambiar Pulso, Organismo o Mind.
type Conexion struct {
	conexion net.Conn
	lector *bufio.Reader
}

func NuevaConexion(conexion net.Conn) *Conexion {
	return &Conexion{
		conexion: conexion,
		lector: bufio.NewReader(conexion),
	}
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

func (conexion *Conexion) Cerrar() error {
	return conexion.conexion.Close()
}
