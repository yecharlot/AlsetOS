# UiRuntime nativo — apps como orges de UI (sin SO anfitrión)

## Qué es

**UiRuntime** es un organismo del kernel (`K_UIRUNTIME`) que interpreta un **IR de interfaz** (árbol de nodos) y lo dibuja dentro de ventanas `K_APP`.

No es un navegador ni HTML. Es el mismo contrato conceptual de **Alset-JS-Runtime / Studio**: componentes, estado, acciones — ejecutado en bare-metal sobre el compositor.

```
Maestro
  └── Desktop
        └── UiRuntime          ← interpreta IR
              └── K_APP        ← ventana cuyo cuerpo es una orge UI
                    state[0..3]  estado reactivo
```

## Por qué existe

| Camino | Rol |
|--------|-----|
| `desktop/shell` + Go | Laboratorio de UX en host (HTML) |
| **kernel + UiRuntime** | Destino: apps Alset en el kernel, sin anfitrión |

Studio/Editor podrán **exportar IR** hacia este runtime. Hoy el IR va **embebido** en `kernel.c` para demos.

## Tipos de nodo (IR)

| Código | Nombre | Qué hace |
|--------|--------|----------|
| `UI_PANEL` | panel | Contenedor con borde |
| `UI_LABEL` | etiqueta | Texto estático |
| `UI_BUTTON` | botón | Click → acción (`INC`, `DEC`, `CLR`, `PULSE`, `NEXT`) |
| `UI_LIST` | lista | Ítems seleccionables |
| `UI_BADGE` | badge | Muestra `state[slot]` |
| `UI_FIELD` | campo | Texto editable (preparado) |

Cada nodo tiene geometría relativa al **body** de la ventana y puede enlazar un `state_slot`.

## Apps de ejemplo (embebidas)

| App | Icono / menú | Qué prueba |
|-----|--------------|------------|
| **Contador** | icono C · tecla `5` | Estado reactivo +, −, limpiar |
| **Catálogo** | icono A · tecla `6` | Lista + pulso a Desktop |
| **Demo IR** | icono D · tecla `7` | Panel, lista, navegación, pulse |

Menú Inicio también lista estas tres apps IR.

## Cómo probar

```bash
cd kernel_native
mkdir -p "$HOME/tmp" && export TMPDIR="$HOME/tmp"
make run
```

1. Espera el splash o pulsa una tecla.
2. Doble clic en **Contador** (o tecla `5`).
3. Pulsa **+** / **−** / **Limpiar**: el badge cambia (estado de la orge).
4. Abre **Catalogo**: selecciona filas; **Pedir** incrementa pulsos.

## Cómo añadir una app IR nueva

1. Define nodos:

```c
static const struct UiNode app_mia_nodes[] = {
  {UI_LABEL, 0, 12, 8, 200, 20, 15, 0, 0},
  {UI_BUTTON, UI_ACT_INC, 12, 40, 80, 36, 2, 0, 0},
};
```

2. Añade entrada en `ui_apps[]` y una cadena en `ui_str[]` si hace falta.
3. Llama `crear_app(índice)` desde icono, menú o tecla.
4. `make run`.

## Relación con Studio / Alset-JS-Runtime

```
Alset Studio / Editor
        │  export IR (futuro)
        ▼
   UiAppDef + UiNode[]
        │
        ▼
  kernel UiRuntime  →  framebuffer
```

El shell HTML sigue siendo el laboratorio visual. El kernel es el **organismo base** que ejecuta el mismo modelo sin Linux.

## Archivos

| Archivo | Contenido |
|---------|-----------|
| `kernel_native/src/kernel.c` | IR, `crear_app`, `ui_draw_nodes`, `ui_hit`, organismo UiRuntime |
| `docs/LEARN_DESKTOP.md` | Mapa shell ↔ kernel |
| Este documento | Contrato del runtime de orges UI |

## Límites honestos (hoy)

- IR **compilado dentro** del kernel (aún no carga archivos desde disco).
- Tipografía bitmap, sin CSS genérico.
- Subconjunto de componentes (suficiente para demos y el puente Studio→kernel).

El siguiente salto natural: **serializar IR desde Studio** y cargarlo en el kernel (o en un mini-nodo que publique orges por ANS).
