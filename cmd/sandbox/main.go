package main

import (
    "github.com/yecharlot/AlsetOS/sandbox"
    "github.com/yecharlot/AlsetOS/zyrion"
)

func main() {
    sandbox.Mostrar(sandbox.EjecutarSimulacion("sandbox-si", zyrion.Si, true))
    sandbox.Mostrar(sandbox.EjecutarSimulacion("sandbox-incierto", zyrion.Incierto, true))
}
