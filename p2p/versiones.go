package p2p

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/yecharlot/AlsetOS/identidad"
)

const VersionDescriptorSchema = "alset.version.v1"

var (
	ErrVersionDescriptorInvalido = errors.New("descriptor de versión inválido")
	ErrLinajeEnConflicto = errors.New("linaje con versiones en conflicto")
	ErrLinajeIncompleto = errors.New("linaje incompleto: falta un antecesor")
)

// DescriptorVersion registra una versión inmutable de un organismo lógico.
// OrganismoID es una clave estable elegida por el dominio; RootCID identifica
// exclusivamente los bytes de esta versión. La firma prueba autoría, no permiso.
type DescriptorVersion struct {
	Schema       string    `json:"schema"`
	OrganismoID  string    `json:"organismo_id"`
	Revision     uint64    `json:"revision"`
	RootCID      string    `json:"root_cid"`
	Anterior     string    `json:"anterior,omitempty"`
	AutorNodeID  string    `json:"autor_node_id"`
	AutorPeerID  string    `json:"autor_peer_id"`
	ClavePublica string    `json:"clave_publica"`
	Fecha        time.Time `json:"fecha"`
	Firma        string    `json:"firma"`
}

type descriptorFirmable struct {
	Schema       string    `json:"schema"`
	OrganismoID  string    `json:"organismo_id"`
	Revision     uint64    `json:"revision"`
	RootCID      string    `json:"root_cid"`
	Anterior     string    `json:"anterior,omitempty"`
	AutorNodeID  string    `json:"autor_node_id"`
	AutorPeerID  string    `json:"autor_peer_id"`
	ClavePublica string    `json:"clave_publica"`
	Fecha        time.Time `json:"fecha"`
}

type EstadoLinaje string

const (
	LinajeCoherente EstadoLinaje = "coherente"
	LinajeIncompleto EstadoLinaje = "incompleto"
	LinajeConflicto EstadoLinaje = "conflicto"
)

type IndiceVersiones struct {
	mu sync.RWMutex
	ruta string
	datos map[string][]DescriptorVersion
}

func CrearDescriptorVersion(id *identidad.Identidad, organismoID string, revision uint64, rootCID, anterior string) (DescriptorVersion, error) {
	if id == nil || len(id.ClavePrivada) != ed25519.PrivateKeySize || len(id.ClavePublica) != ed25519.PublicKeySize {
		return DescriptorVersion{}, fmt.Errorf("%w: identidad ausente o inválida", ErrVersionDescriptorInvalido)
	}
	d := DescriptorVersion{
		Schema: VersionDescriptorSchema,
		OrganismoID: strings.TrimSpace(organismoID),
		Revision: revision,
		RootCID: strings.TrimSpace(rootCID),
		Anterior: strings.TrimSpace(anterior),
		AutorNodeID: id.ID,
		AutorPeerID: peerIDDeClave(id.ClavePublica),
		ClavePublica: base64.StdEncoding.EncodeToString(id.ClavePublica),
		Fecha: time.Now().UTC(),
	}
	if err := validarCamposDescriptor(d); err != nil { return DescriptorVersion{}, err }
	firmable := d.firmable()
	bytes, err := json.Marshal(firmable)
	if err != nil { return DescriptorVersion{}, fmt.Errorf("serializar descriptor: %w", err) }
	d.Firma = base64.StdEncoding.EncodeToString(id.Firmar(bytes))
	return d, nil
}

func (d DescriptorVersion) firmable() descriptorFirmable {
	return descriptorFirmable{d.Schema,d.OrganismoID,d.Revision,d.RootCID,d.Anterior,d.AutorNodeID,d.AutorPeerID,d.ClavePublica,d.Fecha}
}

func validarCamposDescriptor(d DescriptorVersion) error {
	if d.Schema != VersionDescriptorSchema || d.OrganismoID == "" || len(d.OrganismoID) > 256 || d.Revision == 0 {
		return fmt.Errorf("%w: schema, organismo o revisión inválidos", ErrVersionDescriptorInvalido)
	}
	if _, err := CIDRoot(d.RootCID); err != nil { return fmt.Errorf("%w: RootCID mal formado: %v", ErrVersionDescriptorInvalido, err) }
	if d.Revision == 1 && d.Anterior != "" { return fmt.Errorf("%w: revisión inicial no admite antecesor", ErrVersionDescriptorInvalido) }
	if d.Revision > 1 && (d.Anterior == "" || func() bool { _, err := CIDRoot(d.Anterior); return err != nil }()) {
		return fmt.Errorf("%w: revisión posterior requiere RootCID antecesor", ErrVersionDescriptorInvalido)
	}
	if d.Fecha.IsZero() || d.AutorNodeID == "" || d.AutorPeerID == "" || d.ClavePublica == "" {
		return fmt.Errorf("%w: faltan datos de autoría", ErrVersionDescriptorInvalido)
	}
	return nil
}

// VerificarDescriptor comprueba firma, NodeID y PeerID derivados de la misma
// clave Ed25519. La validez criptográfica no concede confianza ni autorización.
func VerificarDescriptor(d DescriptorVersion) error {
	if err := validarCamposDescriptor(d); err != nil { return err }
	publica, err := base64.StdEncoding.DecodeString(d.ClavePublica)
	if err != nil || len(publica) != ed25519.PublicKeySize { return fmt.Errorf("%w: clave pública inválida", ErrVersionDescriptorInvalido) }
	nodeID := "node:" + base64.RawURLEncoding.EncodeToString(publica)
	if d.AutorNodeID != nodeID { return fmt.Errorf("%w: NodeID no corresponde a la clave", ErrVersionDescriptorInvalido) }
	if d.AutorPeerID != peerIDDeClave(ed25519.PublicKey(publica)) { return fmt.Errorf("%w: PeerID no corresponde a la clave", ErrVersionDescriptorInvalido) }
	firma, err := base64.StdEncoding.DecodeString(d.Firma)
	if err != nil || len(firma) != ed25519.SignatureSize { return fmt.Errorf("%w: firma inválida", ErrVersionDescriptorInvalido) }
	contenido, err := json.Marshal(d.firmable())
	if err != nil { return err }
	if !ed25519.Verify(ed25519.PublicKey(publica), contenido, firma) { return fmt.Errorf("%w: firma no válida", ErrVersionDescriptorInvalido) }
	return nil
}

func peerIDDeClave(publica ed25519.PublicKey) string {
	clave, err := libp2pcrypto.UnmarshalEd25519PublicKey(publica)
	if err != nil { return "" }
	id, err := peer.IDFromPublicKey(clave)
	if err != nil { return "" }
	return id.String()
}

func NuevoIndiceVersiones(ruta string) (*IndiceVersiones, error) {
	i := &IndiceVersiones{ruta:ruta, datos:make(map[string][]DescriptorVersion)}
	if ruta == "" { return i, nil }
	bytes, err := os.ReadFile(ruta)
	if os.IsNotExist(err) { return i, nil }
	if err != nil { return nil, fmt.Errorf("leer índice de versiones: %w", err) }
	if len(bytes) == 0 { return i, nil }
	if err := json.Unmarshal(bytes, &i.datos); err != nil { return nil, fmt.Errorf("decodificar índice de versiones: %w", err) }
	if i.datos == nil { i.datos = make(map[string][]DescriptorVersion) }
	for organismo, versiones := range i.datos {
		for _, d := range versiones {
			if d.OrganismoID != organismo { return nil, fmt.Errorf("%w: clave de índice no coincide", ErrVersionDescriptorInvalido) }
			if err := VerificarDescriptor(d); err != nil { return nil, err }
		}
	}
	return i, nil
}

// Registrar conserva también las versiones en conflicto: no elige ganadores.
// La clasificación se calcula sobre el conjunto completo de evidencias.
func (i *IndiceVersiones) Registrar(d DescriptorVersion) (EstadoLinaje, error) {
	if i == nil { return "", fmt.Errorf("índice de versiones nulo") }
	if err := VerificarDescriptor(d); err != nil { return "", err }
	i.mu.Lock()
	defer i.mu.Unlock()
	actual := append([]DescriptorVersion(nil), i.datos[d.OrganismoID]...)
	for _, existente := range actual {
		if existente.Revision == d.Revision && existente.RootCID == d.RootCID && existente.Anterior == d.Anterior && existente.Firma == d.Firma {
			return clasificarLinaje(actual), nil
		}
	}
	nuevo := append(actual, d)
	sort.Slice(nuevo, func(a,b int) bool {
		if nuevo[a].Revision == nuevo[b].Revision { return nuevo[a].RootCID < nuevo[b].RootCID }
		return nuevo[a].Revision < nuevo[b].Revision
	})
	anterior := i.datos[d.OrganismoID]
	i.datos[d.OrganismoID] = nuevo
	if err := i.persistirLocked(); err != nil {
		i.datos[d.OrganismoID] = anterior
		return "", err
	}
	return clasificarLinaje(nuevo), nil
}

// Todas devuelve una instantánea determinista del inventario local.
// Los descriptores siguen siendo evidencia firmada; el inventario no otorga confianza.
func (i *IndiceVersiones) Todas() []DescriptorVersion {
	if i == nil { return nil }
	i.mu.RLock()
	defer i.mu.RUnlock()
	organismos := make([]string, 0, len(i.datos))
	for organismo := range i.datos { organismos = append(organismos, organismo) }
	sort.Strings(organismos)
	resultado := make([]DescriptorVersion, 0)
	for _, organismo := range organismos {
		resultado = append(resultado, i.datos[organismo]...)
	}
	return resultado
}

func (i *IndiceVersiones) Versiones(organismoID string) []DescriptorVersion {
	if i == nil { return nil }
	i.mu.RLock(); defer i.mu.RUnlock()
	return append([]DescriptorVersion(nil), i.datos[organismoID]...)
}

func (i *IndiceVersiones) Estado(organismoID string) EstadoLinaje {
	return clasificarLinaje(i.Versiones(organismoID))
}

func clasificarLinaje(versiones []DescriptorVersion) EstadoLinaje {
	if len(versiones) == 0 { return LinajeIncompleto }
	porRevision := make(map[uint64]map[string]struct{})
	porAntecesor := make(map[string]map[string]struct{})
	padresPorVersion := make(map[string]map[string]struct{})
	presentes := make(map[string]map[uint64]struct{})
	for _, d := range versiones {
		if porRevision[d.Revision] == nil { porRevision[d.Revision] = make(map[string]struct{}) }
		porRevision[d.Revision][d.RootCID] = struct{}{}
		claveVersion := fmt.Sprintf("%d:%s", d.Revision, d.RootCID)
		if padresPorVersion[claveVersion] == nil { padresPorVersion[claveVersion] = make(map[string]struct{}) }
		padresPorVersion[claveVersion][d.Anterior] = struct{}{}
		if d.Anterior != "" {
			if porAntecesor[d.Anterior] == nil { porAntecesor[d.Anterior] = make(map[string]struct{}) }
			porAntecesor[d.Anterior][d.RootCID] = struct{}{}
		}
		if presentes[d.RootCID] == nil { presentes[d.RootCID] = make(map[uint64]struct{}) }
		presentes[d.RootCID][d.Revision] = struct{}{}
	}
	for _, raices := range porRevision { if len(raices) > 1 { return LinajeConflicto } }
	for _, hijos := range porAntecesor { if len(hijos) > 1 { return LinajeConflicto } }
	for _, padres := range padresPorVersion { if len(padres) > 1 { return LinajeConflicto } }
	for _, d := range versiones {
		if d.Revision > 1 {
			if _, ok := presentes[d.Anterior][d.Revision-1]; !ok { return LinajeIncompleto }
		}
	}
	return LinajeCoherente
}

func (i *IndiceVersiones) persistirLocked() error {
	if i.ruta == "" { return nil }
	contenido, err := json.MarshalIndent(i.datos, "", "  ")
	if err != nil { return fmt.Errorf("serializar índice de versiones: %w", err) }
	if err := os.MkdirAll(dirRuta(i.ruta), 0700); err != nil { return err }
	tmp, err := os.CreateTemp(dirRuta(i.ruta), ".versiones-*")
	if err != nil { return err }
	nombre := tmp.Name()
	ok := false
	defer func(){ if !ok { _ = os.Remove(nombre) } }()
	if err := tmp.Chmod(0600); err != nil { _ = tmp.Close(); return err }
	if _, err := tmp.Write(contenido); err != nil { _ = tmp.Close(); return err }
	if err := tmp.Sync(); err != nil { _ = tmp.Close(); return err }
	if err := tmp.Close(); err != nil { return err }
	if err := os.Rename(nombre, i.ruta); err != nil { return err }
	if d, err := os.Open(dirRuta(i.ruta)); err == nil { _ = d.Sync(); _ = d.Close() }
	ok = true
	return nil
}
