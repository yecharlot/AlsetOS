package recuperacion

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/motor"
	"github.com/yecharlot/AlsetOS/p2p"
)

type Servicio struct {
	Nodo  *p2p.Nodo
	Motor *motor.Motor
}

func Nuevo(nodo *p2p.Nodo) *Servicio {
	return &Servicio{Nodo: nodo, Motor: motor.Nuevo()}
}

func (servicio *Servicio) EjecutarRemoto(ctx context.Context, rootCID string) (motor.Resultado, error) {
	if servicio == nil || servicio.Nodo == nil {
		return motor.Resultado{}, fmt.Errorf("servicio de recuperación sin nodo")
	}

	contenido, err := servicio.Nodo.RecuperarOrganismo(ctx, rootCID)
	if err != nil {
		return motor.Resultado{}, err
	}

	var documento manifiesto.Manifiesto
	if err := json.Unmarshal(contenido, &documento); err != nil {
		return motor.Resultado{}, fmt.Errorf("decodificar organismo recuperado: %w", err)
	}
	if err := documento.Validar(); err != nil {
		return motor.Resultado{}, fmt.Errorf("validar organismo recuperado: %w", err)
	}

	resultado, err := servicio.Motor.EjecutarManifiesto(documento)
	if err != nil {
		return motor.Resultado{}, fmt.Errorf("ejecutar organismo recuperado: %w", err)
	}
	if resultado.RootCID != rootCID {
		return motor.Resultado{}, fmt.Errorf("RootCID recuperado distinto: esperado %s, obtenido %s", rootCID, resultado.RootCID)
	}

	return resultado, nil
}
