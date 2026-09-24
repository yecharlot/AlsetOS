package motor

import (
	"fmt"
	"time"
	"github.com/yecharlot/AlsetOS/gene"
	"github.com/yecharlot/AlsetOS/lispai"
	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/mind"
	"github.com/yecharlot/AlsetOS/organismo"
	"github.com/yecharlot/AlsetOS/pulso"
	"github.com/yecharlot/AlsetOS/rootcid"
)

type Resultado struct { Nombre string; RootCID string; Estado string; Decision string; GenesEjecutados []string; UltimoPulso string }
type Motor struct { RegistroGenes *gene.Registro }

func Nuevo() *Motor {
	registro := gene.NuevoRegistro()
	registro.Registrar(gene.Gene{Nombre: "gene-saludo"})
	registro.Registrar(gene.Gene{Nombre: "gene-sandbox"})
	return &Motor{RegistroGenes: registro}
}

func (motor *Motor) EjecutarManifiesto(documento manifiesto.Manifiesto) (Resultado, error) {
	estado, err := documento.EstadoZyrion()
	if err != nil { return Resultado{}, err }
	entidad := organismo.Nuevo(documento.Nombre)
	entidad.RootCID = rootcid.Crear(entidad.Nombre)
	for _, capacidad := range documento.Capacidades { entidad.Capacidad[capacidad] = true }
	interprete := lispai.Interprete{}
	for _, programa := range documento.LispAI { interprete.Evaluar(programa, entidad) }
	mente := mind.Mente{}
	estrategia := mind.Detener
	if documento.Zyrion.Estrategia == string(mind.Evaluar) { estrategia = mind.Evaluar }
	decision := mente.DecidirConEstrategia(estado, entidad, estrategia)
	resultado := Resultado{Nombre: entidad.Nombre, RootCID: entidad.RootCID, Estado: estado.Texto(), Decision: decision}
	if decision == "ejecutar_gene" {
		for _, nombreGene := range documento.Genes {
			gen, existe := motor.RegistroGenes.Obtener(nombreGene)
			if !existe { return Resultado{}, fmt.Errorf("Gene no registrado: %s", nombreGene) }
			resultado.GenesEjecutados = append(resultado.GenesEjecutados, gen.Ejecutar(entidad))
		}
	}
	pulso.Emitir(entidad, pulso.Pulso{Tipo: "resultado", Origen: entidad.RootCID, Contenido: decision, Fecha: time.Now()})
	resultado.UltimoPulso = entidad.Memoria["ultimo_pulso"]
	return resultado, nil
}
