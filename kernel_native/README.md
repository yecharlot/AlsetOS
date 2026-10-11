# kernel_native — Desktop Alset + UiRuntime

Identidad tomada de `desktop/shell/shell.css` + comportamiento de `shell.js`,
implementado como **organismos** en framebuffer.

Incluye el organismo **UiRuntime**: apps definidas por un IR de nodos
(label, button, list, badge, panel) — el puente hacia Alset Studio / Alset-JS-Runtime
**sin SO anfitrión**.

## Probar

```bash
mkdir -p "$HOME/tmp" && export TMPDIR="$HOME/tmp"
cd kernel_native && make run
```

## Guías

- **[docs/LEARN_DESKTOP.md](../docs/LEARN_DESKTOP.md)** — mapa shell → kernel
- **[docs/UI_RUNTIME.es.md](../docs/UI_RUNTIME.es.md)** — IR, K_APP, cómo añadir apps

## Controles

| Acción | Cómo |
|--------|------|
| Splash → escritorio | Espera o click |
| Apps clásicas | Iconos Studio / Editor / Archivos / Terminal / Mind |
| **Apps IR (orges UI)** | Iconos **Contador**, **Catalogo**, **Demo IR** |
| Menú inicio | Botón **A** en taskbar o tecla **M** |
| Crear clásico | **1** ventana · **2** texto · **3** lista · **4** inspector |
| Crear IR | **5** Contador · **6** Catálogo · **7** Demo IR |
| Cerrar / minimizar | Botones en titlebar |
| Arrastrar | Barra de título |
| Botones de app IR | Click dentro del cuerpo de la ventana K_APP |

## Organismos nuevos

| Nombre | Kind | Rol |
|--------|------|-----|
| UiRuntime | `K_UIRUNTIME` | Interpreta IR y crea/atiende apps |
| *app\** | `K_APP` | Ventana cuyo cuerpo es una orge UI |

## Archivo principal

`src/kernel.c` — IR embebido, compositor y UiRuntime en un solo lugar legible.
