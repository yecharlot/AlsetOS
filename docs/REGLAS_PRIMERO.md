# Reglas primero — cómo crear apps en AlsetOS

## Idea

**No** escribes el runtime a mano:

```go
// esto NO es el flujo deseado para el creador de apps
rt := frontera.NuevoRuntime()
app.Instalar(...)
```

**Sí** escribes **reglas** (manifiesto JSON de organismos, contratos, ARL).  
La **plataforma** ensambla, evalúa Zyrion/ARL, autoriza por capacidades y **puede generar código** como artefacto.

```
reglas.json  →  alset-plataforma aplicar  →  organismos vivos + (opcional) código generado
```

El código generado se puede retocar; la fuente canónica sigue siendo el JSON.

## Comando

```bash
go run ./cmd/alset-plataforma aplicar ejemplos/reglas/cafeteria.json \
  -generar /tmp/cafeteria-gen \
  -data /tmp/cafeteria-data
```

## Qué va en las reglas

- `app`, `version`, `descripcion`
- `organismos[]` con:
  - identidad (`oid`, `name`, `version`, `kind`)
  - `capabilities`, `goals`
  - `contracts` (provides / requires)
  - `initial_state`
  - `rules` (ARL estructurado: when → then)

## Separación de poderes

| Pieza | Rol |
|-------|-----|
| Reglas / ARL | Qué debe ocurrir |
| Zyrion | Si / No / Incierto |
| Capacidades | Autorización |
| Código generado | Refinamiento opcional, no origen |
| Frontera / kernel | Seguridad del nodo (infraestructura) |

## Ejemplo

Ver `ejemplos/reglas/cafeteria.json` (Stock + Venta).
