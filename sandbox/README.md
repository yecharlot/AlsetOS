# Sandbox local de AlsetOS

El sandbox permite ejecutar el núcleo conceptual de AlsetOS completamente en memoria.

No necesita red, GitHub, libp2p, base de datos ni TinyCore.

## Ejecutar

```bash
go run ./cmd/sandbox
```

## Escenarios

- `sandbox-si`: Zyrion permite la ejecución y Mind activa el Gene.
- `sandbox-incierto`: Zyrion produce incertidumbre y Mind detiene el Gene.

El objetivo es disponer de un entorno determinista y barato para experimentar con Organismos antes de introducir infraestructura distribuida.
