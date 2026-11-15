package p2p

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	TopicAnunciosVersion = "alset/versions/1.0.0"
	SchemaAnuncioVersion = "alset.version-announcement.v1"
	maxAnuncioVersion    = 64 << 10
)

type AnuncioVersion struct {
	Schema     string           `json:"schema"`
	Descriptor DescriptorVersion `json:"descriptor"`
}

func codificarAnuncioVersion(d DescriptorVersion) ([]byte, error) {
	if err := VerificarDescriptor(d); err != nil {
		return nil, err
	}
	b, err := json.Marshal(AnuncioVersion{Schema: SchemaAnuncioVersion, Descriptor: d})
	if err != nil {
		return nil, fmt.Errorf("serializar anuncio de versión: %w", err)
	}
	if len(b) > maxAnuncioVersion {
		return nil, fmt.Errorf("anuncio supera el límite de %d bytes", maxAnuncioVersion)
	}
	return b, nil
}

func decodificarAnuncioVersion(b []byte) (AnuncioVersion, error) {
	if len(b) == 0 || len(b) > maxAnuncioVersion {
		return AnuncioVersion{}, fmt.Errorf("tamaño de anuncio inválido")
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var anuncio AnuncioVersion
	if err := dec.Decode(&anuncio); err != nil {
		return AnuncioVersion{}, fmt.Errorf("anuncio mal formado: %w", err)
	}
	var extra interface{}
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return AnuncioVersion{}, fmt.Errorf("el anuncio contiene datos adicionales")
	}
	if anuncio.Schema != SchemaAnuncioVersion {
		return AnuncioVersion{}, fmt.Errorf("schema de anuncio no soportado")
	}
	if err := VerificarDescriptor(anuncio.Descriptor); err != nil {
		return AnuncioVersion{}, err
	}
	return anuncio, nil
}

func validarMensajeAnuncioVersion(_ context.Context, _ peer.ID, mensaje *pubsub.Message) bool {
	if mensaje == nil {
		return false
	}
	_, err := decodificarAnuncioVersion(mensaje.Data)
	return err == nil
}

// PublicarAnuncioVersion propaga únicamente el descriptor firmado. El evento
// es una pista de descubrimiento: no transfiere el bloque ni demuestra que
// RootCID esté disponible en el emisor o en otro peer.
func (nodo *Nodo) PublicarAnuncioVersion(ctx context.Context, descriptor DescriptorVersion) error {
	if nodo == nil || nodo.topicVersiones == nil {
		return fmt.Errorf("GossipSub de versiones no inicializado")
	}
	datos, err := codificarAnuncioVersion(descriptor)
	if err != nil {
		return err
	}
	if err := nodo.topicVersiones.Publish(ctx, datos); err != nil {
		return fmt.Errorf("publicar anuncio GossipSub: %w", err)
	}
	return nil
}

func (nodo *Nodo) consumirAnunciosVersion(ctx context.Context) {
	defer close(nodo.gossipDone)
	for {
		mensaje, err := nodo.suscripcionVersiones.Next(ctx)
		if err != nil {
			return
		}
		anuncio, err := decodificarAnuncioVersion(mensaje.Data)
		if err != nil {
			continue
		}
		// La validación del topic ya comprobó la firma antes de entregar el
		// mensaje. Registrar es idempotente y conserva forks e incompletitud.
		_, _ = nodo.versiones.Registrar(anuncio.Descriptor)
	}
}

func (nodo *Nodo) iniciarGossipVersiones() error {
	if nodo == nil || nodo.Host == nil {
		return fmt.Errorf("host P2P no inicializado")
	}
	ctx, cancelar := context.WithCancel(context.Background())
	ps, err := pubsub.NewGossipSub(
		ctx,
		nodo.Host,
		pubsub.WithMessageSigning(true),
		pubsub.WithStrictSignatureVerification(true),
		pubsub.WithMaxMessageSize(maxAnuncioVersion),
	)
	if err != nil {
		cancelar()
		return fmt.Errorf("crear GossipSub: %w", err)
	}
	if err := ps.RegisterTopicValidator(TopicAnunciosVersion, pubsub.ValidatorEx(validarMensajeAnuncioVersion), pubsub.WithValidatorInline(true)); err != nil {
		cancelar()
		return fmt.Errorf("registrar validador de anuncios: %w", err)
	}
	topic, err := ps.Join(TopicAnunciosVersion)
	if err != nil {
		cancelar()
		return fmt.Errorf("unirse al topic de versiones: %w", err)
	}
	sub, err := topic.Subscribe()
	if err != nil {
		cancelar()
		_ = topic.Close()
		return fmt.Errorf("suscribirse al topic de versiones: %w", err)
	}
	nodo.gossip = ps
	nodo.topicVersiones = topic
	nodo.suscripcionVersiones = sub
	nodo.cancelarGossip = cancelar
	nodo.gossipDone = make(chan struct{})
	go nodo.consumirAnunciosVersion(ctx)
	return nil
}

// EsperarMallaVersiones espera a que el topic tenga al menos el número de
// peers solicitado. No descubre identidades ni concede confianza.
func (nodo *Nodo) EsperarMallaVersiones(ctx context.Context, minimo int) error {
	if nodo == nil || nodo.topicVersiones == nil {
		return fmt.Errorf("GossipSub de versiones no inicializado")
	}
	if minimo < 0 {
		return fmt.Errorf("mínimo de peers inválido")
	}
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if len(nodo.topicVersiones.ListPeers()) >= minimo {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
