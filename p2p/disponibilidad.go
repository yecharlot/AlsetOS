package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yecharlot/AlsetOS/rootcid"
)

const (
	ProtocoloDisponibilidad = "/alset/content-availability/1.0.0"
	schemaConsultaDisponibilidad = "alset.content-availability.request.v1"
	schemaReciboDisponibilidad = "alset.content-availability.receipt.v1"
	maxEdadReciboDisponibilidad = 2 * time.Minute
)

type consultaDisponibilidad struct {
	Schema string `json:"schema"`
	RootCID string `json:"root_cid"`
	Nonce string `json:"nonce"`
}

type ReciboDisponibilidad struct {
	Schema string `json:"schema"`
	RootCID string `json:"root_cid"`
	PeerID string `json:"peer_id"`
	ClavePublica string `json:"clave_publica"`
	Nonce string `json:"nonce"`
	Emitido time.Time `json:"emitido"`
	Expira time.Time `json:"expira"`
	Firma string `json:"firma"`
}

type reciboDisponibilidadFirmable struct {
	Schema string `json:"schema"`
	RootCID string `json:"root_cid"`
	PeerID string `json:"peer_id"`
	ClavePublica string `json:"clave_publica"`
	Nonce string `json:"nonce"`
	Emitido time.Time `json:"emitido"`
	Expira time.Time `json:"expira"`
}

func (r ReciboDisponibilidad) firmable() reciboDisponibilidadFirmable {
	return reciboDisponibilidadFirmable{r.Schema,r.RootCID,r.PeerID,r.ClavePublica,r.Nonce,r.Emitido,r.Expira}
}

// verificarReciboDisponibilidad valida prueba de posesión reciente, ligada a
// un desafío del solicitante. No es una garantía de retención futura.
func verificarReciboDisponibilidad(r ReciboDisponibilidad, root, nonce string, esperado peer.ID, ahora time.Time) error {
	if r.Schema != schemaReciboDisponibilidad || r.RootCID != root || r.Nonce != nonce || r.PeerID != esperado.String() {
		return fmt.Errorf("recibo de disponibilidad no corresponde a la consulta")
	}
	if r.Emitido.After(ahora.Add(30*time.Second)) || !r.Expira.After(ahora) || r.Expira.Sub(r.Emitido) > maxEdadReciboDisponibilidad || ahora.Sub(r.Emitido) > maxEdadReciboDisponibilidad {
		return fmt.Errorf("recibo de disponibilidad vencido o con fecha inválida")
	}
	publica, err := base64.StdEncoding.DecodeString(r.ClavePublica)
	if err != nil || len(publica) != ed25519.PublicKeySize || peerIDDeClave(ed25519.PublicKey(publica)) != r.PeerID {
		return fmt.Errorf("identidad del recibo no corresponde a su clave")
	}
	firma, err := base64.StdEncoding.DecodeString(r.Firma)
	if err != nil || len(firma) != ed25519.SignatureSize {
		return fmt.Errorf("firma del recibo inválida")
	}
	payload, err := json.Marshal(r.firmable())
	if err != nil { return err }
	if !ed25519.Verify(ed25519.PublicKey(publica), payload, firma) {
		return fmt.Errorf("firma del recibo no válida")
	}
	return nil
}

func (nodo *Nodo) recibirConsultaDisponibilidad(stream network.Stream) {
	defer stream.Close()
	if nodo == nil || nodo.Identidad == nil || nodo.almacen == nil {
		_ = json.NewEncoder(stream).Encode(map[string]string{"error":"servicio de disponibilidad no inicializado"})
		return
	}
	limitado := &io.LimitedReader{R:stream,N:4097}
	var consulta consultaDisponibilidad
	if err := json.NewDecoder(bufio.NewReader(limitado)).Decode(&consulta); err != nil || limitado.N <= 0 || consulta.Schema != schemaConsultaDisponibilidad || len(consulta.Nonce) < 32 || len(consulta.Nonce) > 128 {
		_ = json.NewEncoder(stream).Encode(map[string]string{"error":"consulta de disponibilidad inválida"})
		return
	}
	if _, err := CIDRoot(consulta.RootCID); err != nil {
		_ = json.NewEncoder(stream).Encode(map[string]string{"error":"RootCID inválido"})
		return
	}
	disponible, err := nodo.almacen.Disponible(consulta.RootCID)
	if err != nil || !disponible {
		_ = json.NewEncoder(stream).Encode(map[string]string{"error":"contenido no disponible o corrupto"})
		return
	}
	ahora := time.Now().UTC()
	recibo := ReciboDisponibilidad{
		Schema:schemaReciboDisponibilidad,
		RootCID:consulta.RootCID,
		PeerID:nodo.ID().String(),
		ClavePublica:base64.StdEncoding.EncodeToString(nodo.Identidad.ClavePublica),
		Nonce:consulta.Nonce,
		Emitido:ahora,
		Expira:ahora.Add(maxEdadReciboDisponibilidad),
	}
	payload, err := json.Marshal(recibo.firmable())
	if err != nil {
		_ = json.NewEncoder(stream).Encode(map[string]string{"error":"no se pudo firmar el recibo"})
		return
	}
	recibo.Firma = base64.StdEncoding.EncodeToString(nodo.Identidad.Firmar(payload))
	_ = json.NewEncoder(stream).Encode(recibo)
}

// ConsultarDisponibilidad obtiene una prueba fresca y firmada de que un peer
// conserva localmente un bloque verificable. La prueba no garantiza retención.
func (nodo *Nodo) ConsultarDisponibilidad(ctx context.Context, remoto peer.ID, root string) (ReciboDisponibilidad, error) {
	if nodo == nil || nodo.Host == nil { return ReciboDisponibilidad{}, fmt.Errorf("nodo no inicializado") }
	if remoto == "" || remoto == nodo.ID() { return ReciboDisponibilidad{}, fmt.Errorf("peer remoto inválido") }
	if _, err := CIDRoot(root); err != nil { return ReciboDisponibilidad{}, err }
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil { return ReciboDisponibilidad{}, err }
	nonceTexto := base64.RawURLEncoding.EncodeToString(nonce)
	stream, err := nodo.Host.NewStream(ctx, remoto, ProtocoloDisponibilidad)
	if err != nil { return ReciboDisponibilidad{}, fmt.Errorf("consultar disponibilidad en %s: %w", remoto, err) }
	defer stream.Close()
	if err := json.NewEncoder(stream).Encode(consultaDisponibilidad{Schema:schemaSolicitudDisponibilidad,RootCID:root,Nonce:nonceTexto}); err != nil { return ReciboDisponibilidad{}, err }
	limitado := &io.LimitedReader{R:stream,N:16<<10}
	var recibo ReciboDisponibilidad
	if err := json.NewDecoder(bufio.NewReader(limitado)).Decode(&recibo); err != nil { return ReciboDisponibilidad{}, fmt.Errorf("leer recibo de disponibilidad: %w", err) }
	if limitado.N <= 0 { return ReciboDisponibilidad{}, fmt.Errorf("recibo supera el límite") }
	if err := verificarReciboDisponibilidad(recibo,root,nonceTexto,remoto,time.Now().UTC()); err != nil { return ReciboDisponibilidad{}, err }
	return recibo,nil
}

type ResultadoReplicaVerificada struct {
	PeerID string `json:"peer_id"`
	Confirmada bool `json:"confirmada"`
	Error string `json:"error,omitempty"`
}

type EstadoRedundancia struct {
	RootCID string `json:"root_cid"`
	FactorSolicitado int `json:"factor_solicitado"`
	FactorVerificado int `json:"factor_verificado"`
	Resultados []ResultadoReplicaVerificada `json:"resultados"`
}

// AsegurarRedundancia cuenta la copia local y obtiene confirmaciones recientes
// de peers conectados. Replica de forma explícita; nunca selecciona una versión
// por fecha ni resuelve conflictos. Factor incluye la copia local.
func (nodo *Nodo) AsegurarRedundancia(ctx context.Context, descriptor DescriptorVersion, contenido []byte, factor int) (EstadoRedundancia, error) {
	estado := EstadoRedundancia{RootCID:descriptor.RootCID,FactorSolicitado:factor,Resultados:[]ResultadoReplicaVerificada{}}
	if nodo == nil || nodo.Host == nil || nodo.almacen == nil { return estado,fmt.Errorf("servicios de contenido no inicializados") }
	if factor < 1 || factor > 32 { return estado,fmt.Errorf("factor de redundancia fuera del rango 1..32") }
	if err := VerificarDescriptor(descriptor); err != nil { return estado,err }
	if err := rootcid.Verificar(contenido,descriptor.RootCID); err != nil { return estado,err }
	local, err := nodo.almacen.Disponible(descriptor.RootCID)
	if err != nil || !local { return estado,fmt.Errorf("el nodo coordinador no conserva el bloque RootCID") }
	estado.FactorVerificado = 1
	if factor == 1 { return estado,nil }
	peers := nodo.Host.Network().Peers()
	sort.Slice(peers,func(i,j int) bool { return peers[i].String() < peers[j].String() })
	for _, remoto := range peers {
		if estado.FactorVerificado >= factor { break }
		// Una respuesta de replicación puede informar linaje incompleto/conflicto
		// aunque el bloque se haya guardado. Solo el recibo fresco cuenta como copia.
		_, replicaErr := nodo.ReplicarDescriptor(ctx,remoto,descriptor,contenido)
		_, reciboErr := nodo.ConsultarDisponibilidad(ctx,remoto,descriptor.RootCID)
		resultado := ResultadoReplicaVerificada{PeerID:remoto.String()}
		if reciboErr == nil {
			resultado.Confirmada = true
			estado.FactorVerificado++
		} else if replicaErr != nil {
			resultado.Error = replicaErr.Error()+"; verificación: "+reciboErr.Error()
		} else {
			resultado.Error = reciboErr.Error()
		}
		estado.Resultados = append(estado.Resultados,resultado)
	}
	if estado.FactorVerificado < factor {
		return estado,fmt.Errorf("redundancia insuficiente: verificadas %d de %d copias solicitadas",estado.FactorVerificado,factor)
	}
	return estado,nil
}
