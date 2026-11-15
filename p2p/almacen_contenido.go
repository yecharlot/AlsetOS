package p2p

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yecharlot/AlsetOS/rootcid"
)

var ErrContenidoNoDisponible = errors.New("contenido no disponible localmente")

// AlmacenContenido conserva bloques inmutables identificados por el CIDv1 raw
// derivado del RootCID. El layout en disco usa el digest hexadecimal validado;
// nunca se usa una cadena externa como ruta.
type AlmacenContenido struct {
	mu sync.RWMutex
	dir string
	memoria map[string][]byte
}

func NuevoAlmacenContenido(dir string) (*AlmacenContenido, error) {
	if dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("crear almacén de bloques: %w", err)
		}
	}
	return &AlmacenContenido{dir: dir, memoria: make(map[string][]byte)}, nil
}

func rutaBloque(root string) (string, error) {
	if _, err := CIDRoot(root); err != nil { return "", err }
	digest := strings.TrimPrefix(root, "rootcid:")
	return digest, nil
}

// Guardar es idempotente: un RootCID existente nunca se sobrescribe con bytes
// diferentes. Verifica el contenido antes de escribirlo y publica el bloque
// mediante temporal + fsync + rename.
func (a *AlmacenContenido) Guardar(root string, contenido []byte) error {
	if a == nil { return fmt.Errorf("almacén de contenido nulo") }
	if err := rootcid.Verificar(contenido, root); err != nil { return err }
	digest, err := rutaBloque(root)
	if err != nil { return err }

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.dir == "" {
		if existente, ok := a.memoria[root]; ok {
			if !bytes.Equal(existente, contenido) { return fmt.Errorf("colisión de contenido para %s", root) }
			return nil
		}
		a.memoria[root] = append([]byte(nil), contenido...)
		return nil
	}

	dir := filepath.Join(a.dir, digest[:2])
	if err := os.MkdirAll(dir, 0700); err != nil { return fmt.Errorf("crear shard de bloques: %w", err) }
	destino := filepath.Join(dir, digest)
	if existente, err := os.ReadFile(destino); err == nil {
		if err := rootcid.Verificar(existente, root); err != nil { return fmt.Errorf("bloque existente corrupto %s: %w", root, err) }
		if !bytes.Equal(existente, contenido) { return fmt.Errorf("colisión de contenido para %s", root) }
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("leer bloque existente: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".block-*")
	if err != nil { return fmt.Errorf("crear bloque temporal: %w", err) }
	tmpName := tmp.Name()
	ok := false
	defer func() { if !ok { _ = os.Remove(tmpName) } }()
	if err := tmp.Chmod(0600); err != nil { _ = tmp.Close(); return err }
	if _, err := tmp.Write(contenido); err != nil { _ = tmp.Close(); return fmt.Errorf("escribir bloque: %w", err) }
	if err := tmp.Sync(); err != nil { _ = tmp.Close(); return fmt.Errorf("sincronizar bloque: %w", err) }
	if err := tmp.Close(); err != nil { return fmt.Errorf("cerrar bloque: %w", err) }
	if err := os.Rename(tmpName, destino); err != nil { return fmt.Errorf("confirmar bloque: %w", err) }
	if d, err := os.Open(dir); err == nil { _ = d.Sync(); _ = d.Close() }
	ok = true
	return nil
}

// Obtener vuelve a verificar el RootCID al leer. La existencia de un archivo
// nunca se interpreta por sí sola como disponibilidad de contenido válido.
func (a *AlmacenContenido) Obtener(root string) ([]byte, error) {
	if a == nil { return nil, fmt.Errorf("almacén de contenido nulo") }
	digest, err := rutaBloque(root)
	if err != nil { return nil, err }
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.dir == "" {
		contenido, ok := a.memoria[root]
		if !ok { return nil, ErrContenidoNoDisponible }
		copia := append([]byte(nil), contenido...)
		if err := rootcid.Verificar(copia, root); err != nil { return nil, fmt.Errorf("bloque en memoria corrupto: %w", err) }
		return copia, nil
	}
	contenido, err := os.ReadFile(filepath.Join(a.dir, digest[:2], digest))
	if os.IsNotExist(err) { return nil, ErrContenidoNoDisponible }
	if err != nil { return nil, fmt.Errorf("leer bloque %s: %w", root, err) }
	if err := rootcid.Verificar(contenido, root); err != nil { return nil, fmt.Errorf("bloque corrupto %s: %w", root, err) }
	return contenido, nil
}

func (a *AlmacenContenido) Disponible(root string) (bool, error) {
	_, err := a.Obtener(root)
	if errors.Is(err, ErrContenidoNoDisponible) { return false, nil }
	if err != nil { return false, err }
	return true, nil
}


// Reparar restaura un bloque ausente o corrupto solo cuando los bytes nuevos
// verifican el RootCID esperado. Un bloque ya íntegro nunca se sobrescribe.
func (a *AlmacenContenido) Reparar(root string, contenido []byte) error {
	if a == nil { return fmt.Errorf("almacén de contenido nulo") }
	if err := rootcid.Verificar(contenido, root); err != nil { return err }
	digest, err := rutaBloque(root)
	if err != nil { return err }
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.dir == "" {
		if actual, ok := a.memoria[root]; ok {
			if err := rootcid.Verificar(actual, root); err == nil {
				if !bytes.Equal(actual, contenido) { return fmt.Errorf("colisión de contenido para %s", root) }
				return nil
			}
		}
		a.memoria[root] = append([]byte(nil), contenido...)
		return nil
	}
	dir := filepath.Join(a.dir, digest[:2])
	if err := os.MkdirAll(dir, 0700); err != nil { return err }
	destino := filepath.Join(dir, digest)
	actual, leerErr := os.ReadFile(destino)
	if leerErr == nil && rootcid.Verificar(actual, root) == nil {
		if !bytes.Equal(actual, contenido) { return fmt.Errorf("colisión de contenido para %s", root) }
		return nil
	}
	if leerErr != nil && !os.IsNotExist(leerErr) { return fmt.Errorf("leer bloque para reparar: %w", leerErr) }
	tmp, err := os.CreateTemp(dir, ".repair-*")
	if err != nil { return err }
	nombre := tmp.Name()
	ok := false
	defer func(){ if !ok { _ = os.Remove(nombre) } }()
	if err := tmp.Chmod(0600); err != nil { _ = tmp.Close(); return err }
	if _, err := tmp.Write(contenido); err != nil { _ = tmp.Close(); return err }
	if err := tmp.Sync(); err != nil { _ = tmp.Close(); return err }
	if err := tmp.Close(); err != nil { return err }
	if err := os.Rename(nombre, destino); err != nil { return err }
	if d, err := os.Open(dir); err == nil { _ = d.Sync(); _ = d.Close() }
	ok = true
	return nil
}
