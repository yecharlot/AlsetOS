package nodo

import (
	"fmt"
	"sync"
	"time"

	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/motor"
	"github.com/yecharlot/AlsetOS/registro"
)

type Estado string

const (
	Detenido   Estado = "detenido"
	Ejecutando Estado = "ejecutando"
	Error      Estado = "error"
)

type Nodo struct {
	Nombre  string
	Estado  Estado
	Registro *registro.Registro
	Motor   *motor.Motor
	mu      sync.Mutex
}

type Resultado struct {
	Organismo motor.Resultado
	Recuperado bool
}

func Nuevo(nombre, rutaRegistro string) (*Nodo, error) {
	reg, err := registro.Nuevo(rutaRegistro)
	if err != nil {
		return nil, err
	}
	return &Nodo{
		Nombre:   nombre,
		Estado:   Detenido,
		Registro: reg,
		Motor:    motor.Nuevo(),
	}, nil
}

func (nodo *Nodo) Ejecutar(documento manifiesto.Manifiesto) (Resultado, error) {
	nodo.mu.Lock()
	nodo.Estado = Ejecutando
	nodo.mu.Unlock()

	resultado, err := nodo.Motor.EjecutarManifiesto(documento)
	if err != nil {
		nodo.mu.Lock()
		nodo.Estado = Error
		nodo.mu.Unlock()
		return Resultado{}, err
	}

	err = nodo.Registro.Guardar(registro.Entrada{
		Nombre:      resultado.Nombre,
		RootCID:     resultado.RootCID,
		Estado:      "detenido",
		Actualizado: time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		nodo.mu.Lock()
		nodo.Estado = Error
		nodo.mu.Unlock()
		return Resultado{}, err
	}

	nodo.mu.Lock()
	nodo.Estado = Detenido
	nodo.mu.Unlock()

	return Resultado{Organismo: resultado}, nil
}

func (nodo *Nodo) Recuperar(rootCID string, cargar func() (manifiesto.Manifiesto, error)) (Resultado, error) {
	entrada, existe := nodo.Registro.Obtener(rootCID)
	if !existe {
		return Resultado{}, fmt.Errorf("RootCID no registrado: %s", rootCID)
	}

	documento, err := cargar()
	if err != nil {
		return Resultado{}, err
	}

	resultado, err := nodo.Ejecutar(documento)
	if err != nil {
		return Resultado{}, err
	}
	resultado.Recuperado = resultado.Organismo.RootCID == entrada.RootCID
	return resultado, nil
}
