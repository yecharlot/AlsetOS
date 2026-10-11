# Contrato: ciclo de vida, fundacionales y frontera user/kernel

## Ciclo (único)

Estados: `creado` | `listo` | `ejecutando` | `detenido` | `error`

| Transición | Precondición |
|------------|--------------|
| → listo | Preparar (RootCID presente) o VerificarDefinicion OK |
| → ejecutando | Transición legal + RequisitosParaEjecutar |
| → detenido / error | Según mapa |

**Ejecutando significa:** el organismo está bajo el runtime (Go) o el scheduler del nodo (kernel_native) y puede recibir trabajo. No significa “proceso Unix”.

## Identidad

| Concepto | Significado |
|----------|-------------|
| RootCID presente | Cadena no vacía (`rootcid.Presente`) |
| Definición verificada | `rootcid.Verificar(contenido, cid)` OK + `DefinicionVerificada` |

**MUST NOT** tratar RootCID no vacío como validación canónica.

## Fundacionales

- Solo `organismo.NuevoFundacional(nombre, selladoPor)`.
- Campo `Fundacional bool` + `EspacioKernel` sellados en creación.
- **MUST NOT** usar `Memoria["sistema"]` para eludir requisitos.

## Capacidades

- **Mind** propone decisiones.
- **capacidad.Otorgar(autoridad, objetivo, id)** autoriza (autoridad fundacional o `caps.otorgar`).
- **Gene/WASM** ejecutan solo tras `capacidad.Requerir`.

## Frontera (hito P0)

Paquete `frontera`: aislamiento lógico por `recover`.

- Prueba: pánico de organismo usuario → organismo en `error`, **kernel lógico sigue vivo**, otra app puede correr.
- **No** se afirma aislamiento de hardware/paginación todavía.

## Apps con mínimo esfuerzo

```go
rt := frontera.NuevoRuntime()
maestro := organismo.NuevoFundacional("Maestro", "boot")
app.Instalar(rt, maestro, app.Definicion{Nombre: "mi-app", Version: "1"}, []string{"gene.ejecutar"})
app.Correr(rt, "mi-app", func(o *organismo.Organismo) error { /* lógica */ return nil })
```
