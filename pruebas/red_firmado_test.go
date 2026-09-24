package pruebas

import (
	"net"
	"testing"
	"time"

	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/pulso"
	"github.com/yecharlot/AlsetOS/red"
)

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
	if mensaje.NodoID != identidadNodo.ID { t.Fatalf("identidad incorrecta: %s", mensaje.NodoID) }
	if mensaje.Pulso.Contenido != evento.Contenido { t.Fatal("Pulse incorrecto") }
}

func TestRedRechazaPulseManipulado(t *testing.T) {
	identidadNodo, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }

	evento := pulso.Pulso{Tipo: "resultado", Origen: "rootcid:prueba", Contenido: "ejecutar_gene", Fecha: time.Unix(1, 2)}
	mensaje := red.CrearMensajeFirmado(identidadNodo, evento)
	mensaje.Pulso.Contenido = "manipulado"

	if err := red.VerificarMensaje(mensaje); err == nil {
		t.Fatal("un Pulse manipulado fue aceptado")
	}
}
