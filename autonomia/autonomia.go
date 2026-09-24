package autonomia

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/motor"
	"github.com/yecharlot/AlsetOS/p2p"
	"github.com/yecharlot/AlsetOS/rootcid"
)

type Colocacion struct {
	RootCID     string
	Primario    string
	Replicas    []string
	Actualizado time.Time
}

type Registro struct {
	mu    sync.RWMutex
	ruta  string
	items map[string]Colocacion
}

type Servicio struct {
	Nodo          *p2p.Nodo
	Motor         *motor.Motor
	Registro      *Registro
	Intervalo     time.Duration
	Timeout       time.Duration
	FactorReplica int
}

func Nuevo(nodo *p2p.Nodo, ruta string) (*Servicio, error) {
	if nodo == nil {
		return nil, fmt.Errorf("nodo nulo")
	}
	registro, err := NuevoRegistro(ruta)
	if err != nil {
		return nil, err
	}
	return &Servicio{
		Nodo:          nodo,
		Motor:         motor.Nuevo(),
		Registro:      registro,
		Intervalo:     3 * time.Second,
		Timeout:       2 * time.Second,
		FactorReplica: 2,
	}, nil
}

func NuevoRegistro(ruta string) (*Registro, error) {
	registro := &Registro{ruta: ruta, items: make(map[string]Colocacion)}
	if ruta == "" {
		return registro, nil
	}
	contenido, err := os.ReadFile(ruta)
	if os.IsNotExist(err) {
		return registro, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer colocaciones: %w", err)
	}
	if len(contenido) == 0 {
		return registro, nil
	}
	if err := json.Unmarshal(contenido, &registro.items); err != nil {
		return nil, fmt.Errorf("decodificar colocaciones: %w", err)
	}
	return registro, nil
}

func (registro *Registro) Guardar(colocacion Colocacion) error {
	if colocacion.RootCID == "" {
		return fmt.Errorf("RootCID vacío")
	}
	registro.mu.Lock()
	registro.items[colocacion.RootCID] = colocacion
	copia := make(map[string]Colocacion, len(registro.items))
	for root, item := range registro.items {
		item.Replicas = append([]string(nil), item.Replicas...)
		copia[root] = item
	}
	registro.mu.Unlock()
	if registro.ruta == "" {
		return nil
	}
	contenido, err := json.MarshalIndent(copia, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar colocaciones: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(registro.ruta), 0700); err != nil {
		return fmt.Errorf("crear directorio de colocaciones: %w", err)
	}
	temporal := registro.ruta + ".tmp"
	if err := os.WriteFile(temporal, contenido, 0600); err != nil {
		return fmt.Errorf("escribir colocaciones: %w", err)
	}
	if err := os.Rename(temporal, registro.ruta); err != nil {
		return fmt.Errorf("confirmar colocaciones: %w", err)
	}
	return nil
}

func (registro *Registro) Obtener(rootCID string) (Colocacion, bool) {
	registro.mu.RLock()
	defer registro.mu.RUnlock()
	item, ok := registro.items[rootCID]
	item.Replicas = append([]string(nil), item.Replicas...)
	return item, ok
}

func (registro *Registro) Todos() []Colocacion {
	registro.mu.RLock()
	defer registro.mu.RUnlock()
	resultado := make([]Colocacion, 0, len(registro.items))
	for _, item := range registro.items {
		item.Replicas = append([]string(nil), item.Replicas...)
		resultado = append(resultado, item)
	}
	sort.Slice(resultado, func(i, j int) bool {
		return resultado[i].RootCID < resultado[j].RootCID
	})
	return resultado
}

func (servicio *Servicio) Publicar(ctx context.Context, rootCID string, contenido []byte) (Colocacion, error) {
	if rootcid.CrearContenido(contenido) != rootCID {
		return Colocacion{}, fmt.Errorf("contenido no corresponde al RootCID")
	}
	if err := servicio.Nodo.AnunciarOrganismo(ctx, rootCID, contenido); err != nil {
		return Colocacion{}, err
	}

	peers := servicio.Nodo.Conocidos()
	sort.Slice(peers, func(i, j int) bool { return peers[i].ID.String() < peers[j].ID.String() })

	limite := servicio.FactorReplica
	if limite < 0 {
		limite = 0
	}
	if limite > len(peers) {
		limite = len(peers)
	}

	colocacion := Colocacion{
		RootCID:     rootCID,
		Primario:    servicio.Nodo.ID().String(),
		Replicas:    make([]string, 0, limite),
		Actualizado: time.Now().UTC(),
	}

	for _, info := range peers[:limite] {
		if info.ID == servicio.Nodo.ID() {
			continue
		}
		operacionCtx, cancel := context.WithTimeout(ctx, servicio.Timeout)
		err := servicio.Nodo.ReplicarOrganismo(operacionCtx, info.ID, rootCID, contenido)
		cancel()
		if err != nil {
			continue
		}
		colocacion.Replicas = append(colocacion.Replicas, info.ID.String())
	}

	if err := servicio.Registro.Guardar(colocacion); err != nil {
		return Colocacion{}, err
	}
	return colocacion, nil
}

func (servicio *Servicio) LatirUnaVez(ctx context.Context) map[peer.ID]error {
	resultado := make(map[peer.ID]error)
	for _, info := range servicio.Nodo.Conocidos() {
		if info.ID == servicio.Nodo.ID() {
			continue
		}
		pingCtx, cancel := context.WithTimeout(ctx, servicio.Timeout)
		err := servicio.Nodo.EnviarHeartbeat(pingCtx, info.ID)
		cancel()
		resultado[info.ID] = err
		if err == nil {
			servicio.Nodo.RegistrarLatido(info.ID)
		}
	}
	return resultado
}

func (servicio *Servicio) DetectarPerdidos(ahora time.Time) []peer.ID {
	perdidos := make([]peer.ID, 0)
	for _, info := range servicio.Nodo.Conocidos() {
		if servicio.Nodo.EstadoNodo(info.ID, ahora, servicio.Intervalo*2) == "perdido" {
			perdidos = append(perdidos, info.ID)
		}
	}
	return perdidos
}

func (servicio *Servicio) RecuperarOrganismo(ctx context.Context, rootCID string) (motor.Resultado, error) {
	contenido, err := servicio.Nodo.RecuperarOrganismo(ctx, rootCID)
	if err != nil {
		return motor.Resultado{}, err
	}
	if rootcid.CrearContenido(contenido) != rootCID {
		return motor.Resultado{}, fmt.Errorf("integridad RootCID inválida")
	}
	var documento manifiesto.Manifiesto
	if err := json.Unmarshal(contenido, &documento); err != nil {
		return motor.Resultado{}, fmt.Errorf("decodificar organismo: %w", err)
	}
	if err := documento.Validar(); err != nil {
		return motor.Resultado{}, fmt.Errorf("validar organismo: %w", err)
	}
	resultado, err := servicio.Motor.EjecutarManifiesto(documento)
	if err != nil {
		return motor.Resultado{}, fmt.Errorf("reinstanciar organismo: %w", err)
	}
	if resultado.RootCID != rootCID {
		return motor.Resultado{}, fmt.Errorf("RootCID reinstanciado distinto")
	}
	return resultado, nil
}

func (servicio *Servicio) RecuperarPerdidos(ctx context.Context, perdidos []peer.ID) []error {
	var errores []error
	for _, perdido := range perdidos {
		for _, colocacion := range servicio.Registro.Todos() {
			if colocacion.Primario != perdido.String() {
				continue
			}
			recoveryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			_, err := servicio.RecuperarOrganismo(recoveryCtx, colocacion.RootCID)
			cancel()
			if err != nil {
				errores = append(errores, fmt.Errorf("recuperar %s tras pérdida de %s: %w", colocacion.RootCID, perdido, err))
				continue
			}
			colocacion.Primario = servicio.Nodo.ID().String()
			colocacion.Replicas = filtrarReplica(colocacion.Replicas, servicio.Nodo.ID().String())
			colocacion.Actualizado = time.Now().UTC()
			if err := servicio.Registro.Guardar(colocacion); err != nil {
				errores = append(errores, err)
			}
		}
	}
	return errores
}

func filtrarReplica(replicas []string, nodo string) []string {
	resultado := make([]string, 0, len(replicas))
	for _, replica := range replicas {
		if replica != nodo {
			resultado = append(resultado, replica)
		}
	}
	return resultado
}

func (servicio *Servicio) Ejecutar(ctx context.Context) {
	ticker := time.NewTicker(servicio.Intervalo)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			servicio.LatirUnaVez(ctx)
			perdidos := servicio.DetectarPerdidos(time.Now().UTC())
			if len(perdidos) > 0 {
				_ = servicio.RecuperarPerdidos(ctx, perdidos)
			}
		}
	}
}
