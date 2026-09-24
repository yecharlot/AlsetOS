package lispai

import (
	"strings"

	"github.com/yecharlot/AlsetOS/organismo"
)

// Interprete es el intérprete LispAI mínimo del núcleo inicial.
type Interprete struct{}

// Evaluar ejecuta la instrucción inicial (recordar clave valor).
func (interprete Interprete) Evaluar(programa string, entidad *organismo.Organismo) string {
	instruccion := strings.TrimSpace(programa)
	if strings.HasPrefix(instruccion, "(recordar ") && strings.HasSuffix(instruccion, ")") {
		contenido := strings.TrimSuffix(strings.TrimPrefix(instruccion, "(recordar "), ")")
		partes := strings.SplitN(contenido, " ", 2)
		if len(partes) == 2 {
			entidad.Memoria[partes[0]] = strings.Trim(partes[1], """)
			return "memoria actualizada"
		}
	}
	return "sin instrucción"
}
