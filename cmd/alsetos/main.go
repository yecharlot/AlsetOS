package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Zyrion representa la lógica ternaria nativa de AlsetOS.
// 0 = no, 1 = sí, 2 = incierto.
type Zyrion int

const (
	ZyrionNo Zyrion = iota
	ZyrionSi
	ZyrionIncierto
)

func (valor Zyrion) Texto() string {
	switch valor {
	case ZyrionNo:
		return "no"
	case ZyrionSi:
		return "si"
	default:
		return "incierto"
	}
}

// Organismo es la unidad ejecutable persistente de AlsetOS.
type Organismo struct {
	Nombre   string
	RootCID  string
	Memoria  map[string]string
	Capacidad map[string]bool
}

// Pulso es el mecanismo mínimo de comunicación y trazabilidad.
type Pulso struct {
	Tipo      string
	Origen    string
	Contenido string
	Fecha     time.Time
}

// LispAI es un intérprete mínimo para la primera prueba del núcleo.
type LispAI struct{}

func (interprete LispAI) Evaluar(programa string, organismo *Organismo) string {
	// Primera instrucción ejecutable del lenguaje: (recordar clave valor)
	instruccion := strings.TrimSpace(programa)
	if strings.HasPrefix(instruccion, "(recordar ") && strings.HasSuffix(instruccion, ")") {
		contenido := strings.TrimSuffix(strings.TrimPrefix(instruccion, "(recordar "), ")")
		partes := strings.SplitN(contenido, " ", 2)
		if len(partes) == 2 {
			organismo.Memoria[partes[0]] = strings.Trim(partes[1], """)
			return "memoria actualizada"
		}
	}
	return "sin instrucción"
}

// Mind coordina observación, decisión y acción.
type Mind struct{}

func (mente Mind) Decidir(estado Zyrion, organismo *Organismo) string {
	if estado == ZyrionSi && organismo.Capacidad["gene.ejecutar"] {
		return "ejecutar_gene"
	}
	return "detener"
}

// Gene es una capacidad ejecutable.
type Gene struct {
	Nombre string
}

func (gen Gene) Ejecutar(organismo *Organismo) string {
	resultado := fmt.Sprintf("Gene %s ejecutado para %s", gen.Nombre, organismo.Nombre)
	organismo.Memoria["ultimo_resultado"] = resultado
	return resultado
}

func CrearRootCID(nombre string) string {
	resumen := sha256.Sum256([]byte("organismo:" + nombre + ":alsetos:v0.1"))
	return "rootcid:" + hex.EncodeToString(resumen[:])
}

func EmitirPulso(organismo *Organismo, pulso Pulso) {
	organismo.Memoria["ultimo_pulso"] = pulso.Tipo + ":" + pulso.Contenido
	fmt.Printf("[PULSO] %s -> %s | %s\n", pulso.Origen, pulso.Tipo, pulso.Contenido)
}

func EjecutarFlujo() string {
	// 1. ORGANISMO
	organismo := &Organismo{
		Nombre:    "organismo-prueba",
		Memoria:   map[string]string{},
		Capacidad: map[string]bool{"gene.ejecutar": true},
	}

	// 2. RootCID
	organismo.RootCID = CrearRootCID(organismo.Nombre)
	fmt.Printf("[ORGANISMO] %s\n", organismo.Nombre)
	fmt.Printf("[ROOTCID] %s\n", organismo.RootCID)

	// 3. LispAI
	interprete := LispAI{}
	fmt.Printf("[LISPAI] %s\n", interprete.Evaluar("(recordar origen lisPai)", organismo))

	// 4. Zyrion
	estado := ZyrionSi
	fmt.Printf("[ZYRION] estado=%s\n", estado.Texto())

	// 5. Mind
	mente := Mind{}
	decision := mente.Decidir(estado, organismo)
	fmt.Printf("[MIND] decisión=%s\n", decision)

	// 6. Gene
	gen := Gene{Nombre: "gene-saludo"}
	var resultado string
	if decision == "ejecutar_gene" {
		resultado = gen.Ejecutar(organismo)
		fmt.Printf("[GENE] %s\n", resultado)
	}

	// 7. Pulso
	EmitirPulso(organismo, Pulso{
		Tipo:      "resultado",
		Origen:    organismo.RootCID,
		Contenido: resultado,
		Fecha:     time.Now(),
	})

	return resultado
}

func main() {
	resultado := EjecutarFlujo()
	if resultado == "" {
		panic("flujo AlsetOS sin resultado")
	}
	fmt.Printf("[RESULTADO] %s\n", resultado)
}
