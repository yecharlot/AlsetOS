# Panel financiero con LISPAI-AOS

## Idea

El análisis no es solo un informe en consola: es un **organismo de UI** (`ui-finance`) que observa el estado de `balance`, `dre`, `flujo` y `analisis-financiero`.

## Secuencia lógica

1. **Ingestión** — archivo o JSON → `alset.finance.statement/1.0`
2. **ORGES de datos** — `balance`, `dre`, `flujo`
3. **ORGE de análisis** — ratios + reglas de alerta + genes
4. **ORGE UI** — splash, KPIs, tabs, acciones
5. **Assemble + surface** — el runtime pinta `ui-finance`
6. **Deploy** — PWA en `/apps/{nombre}/` sin chrome del Studio

## Piezas visuales usadas

| Elemento AOS | Rol en el panel |
|--------------|-----------------|
| `(splash …)` | Identidad entidad · periodo |
| `(gradient …)` | Marca / atmósfera |
| `(badge (bind …))` | KPI reactivo |
| `(bottom-tabs …)` | Navegación Resumen / Balance / DRE / Flujo / Alertas |
| `(button (emit …))` | Acciones del organismo |

## Patrón recomendado

Declarar **estado y reglas** en ORGES de dominio; la UI solo **bind** y **emit**. Así Mind y los genes pueden evolucionar sin reescribir el HTML.

## Archivos

- Generador: `mind/ingest_finance.go` → `emitFinanceAOS`
- Plantilla: `plantillas/analisis-financiero.aos`
- Studio: ejemplo `analisis-financiero`
