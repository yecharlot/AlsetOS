package pruebas

import (
	"net"
	"testing"
	"time"

	"github.com/yecharlot/AlsetOS/pulso"
	"github.com/yecharlot/AlsetOS/red"
)

func TestRedTransportaPulse(t *testing.T) {
	servidor, cliente := net.Pipe()
	defer servidor.Close()
	defer cliente.Close()

	enviador := red.NuevaConexion(cliente)
	receptor := red.NuevaConexion(servidor)

	evento := pulso.Pulso{
		Tipo: "resultado",
		Origen: "rootcid:prueba",
		Contenido: "ejecutar_gene",
		Fecha: time.Unix(123, 456),
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- enviador.EnviarPulso(evento)
	}()

	recibido, err := receptor.RecibirPulso()
	if err != nil {
		t.Fatal(err)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}

	if recibido.Tipo != evento.Tipo ||
		recibido.Origen != evento.Origen ||
		recibido.Contenido != evento.Contenido ||
		recibido.Fecha != evento.Fecha {
		t.Fatalf("Pulse alterado durante transporte: %#v != %#v", recibido, evento)
	}
}
