package motor

import (
	"fmt"
	"time"

	"github.com/yecharlot/AlsetOS/agente"
	"github.com/yecharlot/AlsetOS/capacidad"
	"github.com/yecharlot/AlsetOS/gene"
	"github.com/yecharlot/AlsetOS/lispai"
	"github.com/yecharlot/AlsetOS/manifiesto"
	"github.com/yecharlot/AlsetOS/memoria"
	"github.com/yecharlot/AlsetOS/mind"
	"github.com/yecharlot/AlsetOS/organismo"
	"github.com/yecharlot/AlsetOS/pulso"
	"github.com/yecharlot/AlsetOS/rootcid"
)

type Resultado struct {
	Nombre string
	RootCID string
	Estado string
	Decision string
	GenesEjecutados []string
	AgentesEjecutados []string
	UltimoPulso string
}

type Motor struct {
	RegistroGenes *gene.Registro
}

func Nuevo() *Motor {
	registro := gene.NuevoRegistro()
	registro.Registrar(gene.Gene{Nombre: "gene-saludo"})
	registro.Registrar(gene.Gene{Nombre: "gene-sandbox"})
	return &Motor{RegistroGenes: registro}
}

func (motor *Motor) EjecutarManifiesto(documento manifiesto.Manifiesto) (Resultado, error) {
	if err := documento.Validar(); err != nil {
		return Resultado{}, err
	}

	estado, err := documento.EstadoZyrion()
	if err != nil {
		return Resultado{}, err
	}

	contenido, err := documento.Canonico()
	if err != nil {
		return Resultado{}, err
	}

	entidad := organismo.Nuevo(documento.Nombre)
	entidad.RootCID = rootcid.CrearContenido(contenido)
	entidad.Estado = organismo.Listo

	if documento.Memoria != "" {
		datos, err := memoria.Cargar(documento.Memoria)
		if err != nil {
			entidad.Estado = organismo.Error
			return Resultado{}, err
		}
		entidad.Memoria = datos
	}

	for _, nombre := range documento.Capacidades {
		entidad.Capacidad[nombre] = true
	}

	interprete := lispai.Interprete{}
	for _, programa := range documento.LispAI {
		interprete.Evaluar(programa, entidad)
	}

	mente := mind.Mente{}
	estrategia := mind.Detener
	if documento.Zyrion.Estrategia == string(mind.Evaluar) {
		estrategia = mind.Evaluar
	}
	decision := mente.DecidirConEstrategia(estado, entidad, estrategia)

	resultado := Resultado{
		Nombre: entidad.Nombre,
		RootCID: entidad.RootCID,
		Estado: estado.Texto(),
		Decision: decision,
	}

	if decision == "ejecutar_gene" {
		if err := capacidad.Requerir(entidad, "gene.ejecutar"); err != nil {
			entidad.Estado = organismo.Error
			return Resultado{}, err
		}
		entidad.Estado = organismo.Ejecutando

		for _, nombreGene := range documento.Genes {
			gen, existe := motor.RegistroGenes.Obtener(nombreGene)
			if !existe {
				entidad.Estado = organismo.Error
				return Resultado{}, fmt.Errorf("Gene no registrado: %s", nombreGene)
			}
			resultado.GenesEjecutados = append(resultado.GenesEjecutados, gen.Ejecutar(entidad))
		}
	}

	for _, nombreAgente := range documento.Agentes {
		if err := capacidad.Requerir(entidad, "agente.ejecutar"); err != nil {
			entidad.Estado = organismo.Error
			return Resultado{}, err
		}
		actor := agente.Agente{Nombre: nombreAgente, Objetivo: decision}
		actor.Actuar(entidad, "procesar:"+decision)
		resultado.AgentesEjecutados = append(resultado.AgentesEjecutados, nombreAgente)
	}

	pulso.Emitir(entidad, pulso.Pulso{
		Tipo: "resultado",
		Origen: entidad.RootCID,
		Contenido: decision,
		Fecha: time.Now(),
	})
	resultado.UltimoPulso = entidad.Memoria["ultimo_pulso"]
	entidad.Estado = organismo.Detenido

	if documento.Memoria != "" {
		if err := memoria.Guardar(documento.Memoria, entidad.Memoria); err != nil {
			entidad.Estado = organismo.Error
			return Resultado{}, err
		}
	}

	return resultado, nil
}
