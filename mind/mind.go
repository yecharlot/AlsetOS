package mind

import (
 "github.com/yecharlot/AlsetOS/organismo"
 "github.com/yecharlot/AlsetOS/zyrion"
)

type EstrategiaIncierto string

const (
 Detener EstrategiaIncierto = "detener"
 Evaluar EstrategiaIncierto = "evaluar"
)

type Mente struct{}

func (mente Mente) Decidir(estado zyrion.Valor, entidad *organismo.Organismo) string {
 return mente.DecidirConEstrategia(estado, entidad, Detener)
}

func (mente Mente) DecidirConEstrategia(estado zyrion.Valor, entidad *organismo.Organismo, estrategia EstrategiaIncierto) string {
 switch estado {
 case zyrion.Si:
  if entidad.Capacidad["gene.ejecutar"] { return "ejecutar_gene" }
  return "detener"
 case zyrion.Incierto:
  if estrategia == Evaluar { return "evaluar_incierto" }
  return "detener"
 default:
  return "detener"
 }
}

func EvaluarIncierto(entidad *organismo.Organismo) string {
 switch entidad.Memoria["evaluacion"] {
 case "si":
  if entidad.Capacidad["gene.ejecutar"] { return "ejecutar_gene" }
 case "no":
  return "detener"
 }
 return "detener"
}
