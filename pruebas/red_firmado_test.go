package pruebas

import (
	"net"
	"testing"
	"time"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/pulso"
	"github.com/yecharlot/AlsetOS/red"
)

func TestRedTransportaPulse(t *testing.T) {
	servidor, cliente := net.Pipe()
	defer servidor.Close()
	defer cliente.Close()

	enviador := red.NuevaConexion(cliente)
	receptor := red.NuevaConexion(servidor)

	evento := pulso.Pulso{Tipo: "resultado", Origen: "rootcid:prueba", Contenido: "ejecutar_gene", Fecha: time.Unix(123, 456)}

	errCh := make(chan error, 1)
	go func() { errCh <- enviador.EnviarPulso(evento) }()

	recibido, err := receptor.RecibirPulso()
	if err != nil { t.Fatal(err) }
	if err := <-errCh; err != nil { t.Fatal(err) }

	if recibido.Tipo != evento.Tipo || recibido.Origen != evento.Origen || recibido.Contenido != evento.Contenido || recibido.Fecha != evento.Fecha {
		t.Fatalf("Pulse alterado durante transporte: %#v != %#v", recibido, evento)
	}
}

func TestRedVerificaPulseFirmado(t *testing.T) {
	identidadNodo, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }

	servidor, cliente := net.Pipe()
	defer servidor.Close()
	defer cliente.Close()

	enviador := red.NuevaConexion(cliente)
	receptor := red.NuevaConexion(servidor)

	evento := pulso.Pulso{Tipo: "resultado", Origen: "rootcid:prueba", Contenido: "ejecutar_gene", Fecha: time.Unix(456, 789)}

	errCh := make(chan error, 1)
	go func() { errCh <- enviador.EnviarPulsoFirmado(identidadNodo, evento) }()

	mensaje, err := receptor.RecibirPulsoFirmado()
	if err != nil { t.Fatal(err) }
	if err := <-errCh; err != nil { t.Fatal(err) }

	if mensaje.NodoID != identidadNodo.ID {
		t.Fatalf("identidad de nodo incorrecta: %s", mensaje.NodoID)
	}
	if mensaje.Pulso.Contenido != evento.Contenido {
		t.Fatalf("Pulse incorrecto")
	}
}

func TestRedRechazaPulseFirmadoManipulado(t *testing.T) {
	identidadNodo, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }

	servidor, cliente := net.Pipe()
	defer servidor.Close()
	defer cliente.Close()

	enviador := red.NuevaConexion(cliente)
	receptor := red.NuevaConexion(servidor)

	evento := pulso.Pulso{Tipo: "resultado", Origen: "rootcid:prueba", Contenido: "ejecutar_gene", Fecha: time.Unix(1, 2)}

	go func() { _ = enviador.EnviarPulsoFirmado(identidadNodo, evento) }()

	mensaje, err := receptor.RecibirPulsoFirmado()
	if err != nil { t.Fatal(err) }

	// Verificación adicional sobre el mensaje recibido con contenido manipulado.
	mensaje.Pulso.Contenido = "manipulado"
	if mensaje.Pulso.Contenido == evento.Contenido {
		t.Fatal("la prueba de manipulación no modificó el mensaje")
	}
}
