package p2p

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/identidad"
	"github.com/yecharlot/AlsetOS/rootcid"
)

func TestDisponibilidadRemotaFirmadaYRedundanciaVerificada(t *testing.T) {
	idA, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	idB, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	nodoA, err := Nuevo(idA, "/ip4/127.0.0.1/tcp/0")
	if err != nil { t.Fatal(err) }
	defer nodoA.Cerrar()
	nodoB, err := Nuevo(idB, "/ip4/127.0.0.1/tcp/0")
	if err != nil { t.Fatal(err) }
	defer nodoB.Cerrar()

	ctx, cancelar := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelar()
	if err := nodoA.Host.Connect(ctx, peer.AddrInfo{ID:nodoB.ID(),Addrs:nodoB.Host.Addrs()}); err != nil { t.Fatal(err) }

	contenido := []byte("bloque para comprobar redundancia")
	root := rootcid.CrearContenido(contenido)
	descriptor, estado, err := nodoA.AnunciarVersion(ctx,"","organismo:disponibilidad",1,root,"",contenido)
	if err != nil || estado != LinajeCoherente { t.Fatalf("anunciar versión: estado=%s err=%v",estado,err) }

	// La consulta no debe declarar disponible un bloque que el peer aún no posee.
	if _, err := nodoA.ConsultarDisponibilidad(ctx,nodoB.ID(),root); err == nil {
		t.Fatal("se aceptó una consulta de disponibilidad contra un nodo sin el bloque")
	}

	resultado, err := nodoA.AsegurarRedundancia(ctx,descriptor,contenido,2)
	if err != nil { t.Fatalf("asegurar redundancia: %+v: %v",resultado,err) }
	if resultado.FactorVerificado != 2 || len(resultado.Resultados) != 1 || !resultado.Resultados[0].Confirmada {
		t.Fatalf("resultado de redundancia inesperado: %+v",resultado)
	}

	recibo, err := nodoA.ConsultarDisponibilidad(ctx,nodoB.ID(),root)
	if err != nil { t.Fatalf("consultar copia remota: %v",err) }
	if err := verificarReciboDisponibilidad(recibo,root,recibo.Nonce,nodoB.ID(),time.Now().UTC()); err != nil {
		t.Fatalf("recibo remoto no verificable: %v",err)
	}
}

func TestReciboDisponibilidadRechazaDesafioYCaducidad(t *testing.T) {
	id, err := identidad.Nueva()
	if err != nil { t.Fatal(err) }
	root := rootcid.CrearContenido([]byte("contenido"))
	descriptor, err := CrearDescriptorVersion(id,"organismo:recibo",1,root,"")
	if err != nil { t.Fatal(err) }
	nonce := "nonce-de-prueba-con-longitud-suficiente"
	recibo := ReciboDisponibilidad{
		Schema:schemaReciboDisponibilidad,
		RootCID:root,
		PeerID:descriptor.AutorPeerID,
		ClavePublica:descriptor.ClavePublica,
		Nonce:nonce,
		Emitido:time.Now().UTC().Add(-4*time.Minute),
		Expira:time.Now().UTC().Add(-2*time.Minute),
	}
	payload, err := json.Marshal(recibo.firmable())
	if err != nil { t.Fatal(err) }
	recibo.Firma = base64.StdEncoding.EncodeToString(id.Firmar(payload))
	remoto, err := peer.Decode(recibo.PeerID)
	if err != nil { t.Fatal(err) }
	if err := verificarReciboDisponibilidad(recibo,root,nonce,remoto,time.Now().UTC()); err == nil {
		t.Fatal("se aceptó un recibo caducado")
	}
	if err := verificarReciboDisponibilidad(recibo,root,"otro-desafio",remoto,time.Now().UTC()); err == nil {
		t.Fatal("se aceptó un recibo ligado a otro desafío")
	}
}
