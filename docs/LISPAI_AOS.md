# LISPAI-AOS — flows, ORGES y UI como organismos

## Qué es

**LISPAI-AOS** es el dialecto textual para declarar **organismos (ORGES)**, **ensamblarlos como LEGO**, **probarlos** y **desplegarlos**. La **UI también es un organismo** (`kind UI`) con IR de superficie.

```
.aos (fuente canónica)
        ↓
  alset-plataforma flow
        ↓
  Runtime declarativo (contratos, ARL/Zyrion, capacidades)
        ↓
  artefactos editables: reglas JSON + ui.surfaces.json + Go
```

No sustituye el kernel UiRuntime nativo: **exporta el mismo contrato conceptual** (column/row/label/button/badge + estado).

## Flujo de trabajo

1. **Escribes** `ejemplos/flows/mostrador.aos`
2. **Pruebas**  
   `go run ./cmd/alset-plataforma flow ejemplos/flows/mostrador.aos --data /tmp/aos-data`
3. **Generas código editable**  
   `… flow … --generar /tmp/aos-gen`
4. **Retocas** JSON/Go/UI si hace falta; la fuente sigue siendo el `.aos` hasta que decidas promover el artefacto.

## Formas principales

| Forma | Rol |
|-------|-----|
| `(flow nombre …)` | App / ensamblaje |
| `(orge nombre …)` | Organismo de dominio o UI |
| `(provides\|requires contrato ver)` | Contratos LEGO |
| `(state (k v)…)` | Estado inicial |
| `(rule id (when …) (then …))` | ARL estructurado |
| `(ui (column …))` | Superficie — UI-orge |
| `(assemble a b c)` | Piezas del flow |
| `(surface oid)` | Superficie activa |

### UI (IR)

`column` · `row` · `panel` · `label` · `button` · `badge` · `field`  
Botones: `(button "Vender" (emit venta.registrar))`  
Badges: `(badge (bind stock.unidades))`

## Relación con el resto de AlsetOS

| Pieza | Rol |
|-------|-----|
| `declarativo.Runtime` | Ensambla y ejecuta reglas |
| `lispai` mínimo | `recordar` / etc. (otro nivel) |
| UiRuntime kernel | Destino bare-metal del mismo IR |
| JSON reglas | Camino equivalente sin texto AOS |

## Límites (honestos, v0)

- No hay IDE visual de flows todavía (Studio se enlazará después).
- `deploy` multiplataforma = persistir + artefactos; empaquetado USB/Windows es capa aparte.
- LISPAI-AOS cubre el **subconjunto** de ATS-001 usado en pedidos/cafetería, no todo el lenguaje aspiracional.
