// Package app: superficie mínima para crear apps/ORGES con poco esfuerzo (estilo Alset).
//
// Flujo declarativo:
//  1. Definir App{Nombre, Definicion}
//  2. Instalar en Runtime (calcula RootCID, verifica, registra)
//  3. Correr handler de usuario tras la frontera
package app

import (
	"fmt"

	"github.com/yecharlot/AlsetOS/capacidad"
	"github.com/yecharlot/AlsetOS/frontera"
	"github.com/yecharlot/AlsetOS/organismo"
	"github.com/yecharlot/AlsetOS/rootcid"
)

// Definicion canónica de una app (contenido que produce RootCID).
type Definicion struct {
	Nombre  string
	Version string
	Metas   []string
}

// Bytes canónicos para RootCID.
func (d Definicion) Bytes() []byte {
	s := "alset-app:" + d.Nombre + ":v:" + d.Version
	for _, m := range d.Metas {
		s += ":meta:" + m
	}
	return []byte(s)
}

// Instalar crea organismo de usuario, verifica definición y registra en frontera.
// autoridad debe poder otorgar caps básicas de ejecución si se piden.
func Instalar(rt *frontera.Runtime, autoridad *organismo.Organismo, def Definicion, caps []string) (*organismo.Organismo, error) {
	if rt == nil {
		return nil, fmt.Errorf("runtime nulo")
	}
	contenido := def.Bytes()
	cid := rootcid.CrearContenido(contenido)
	o := organismo.Nuevo(def.Nombre)
	if err := o.Preparar(cid); err != nil {
		return nil, err
	}
	if err := o.VerificarDefinicion(cid, rootcid.VerificarOK(contenido, cid)); err != nil {
		return nil, err
	}
	for _, c := range caps {
		if err := capacidad.Otorgar(autoridad, o, c); err != nil {
			return nil, fmt.Errorf("otorgar %s: %w", c, err)
		}
	}
	if err := rt.Registrar(o); err != nil {
		return nil, err
	}
	return o, nil
}

// Correr ejecuta la app tras la frontera (aislamiento lógico).
func Correr(rt *frontera.Runtime, nombre string, h frontera.Handler) frontera.Resultado {
	return rt.EjecutarUsuario(nombre, h)
}
