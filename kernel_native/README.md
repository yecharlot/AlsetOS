# AlsetOS Kernel nativo — campo vivo

Todo es organismo. Sin SO anfitrión. Sin imitar Windows ni Linux.

```
Maestro → Framebuffer → Compositor → Desktop → Orges
```

## Identidad visual (soberana)

- Fondo en degradado profundo + **orbe de presencia** (no rejilla tipo juego).
- **Dock flotante** centrado (no barra de sistema clásica).
- Ventanas con borde fino y **línea de acento** solo en foco.
- Top bar mínima: marca AlsetOS · campo vivo + contadores de organismos/pulsos.

## Rendimiento del ratón

El movimiento del cursor **ya no redibuja toda la escena**.

1. La escena se guarda en un buffer `scene`.
2. Al mover el ratón solo se restaura el rectángulo del cursor y se pinta el nuevo.
3. La escena completa se recalcula solo en click, arrastre o cambio de UI.

Eso es lo que quita la sensación de “emulador de los 80 lento”.

## Arranque

```bash
mkdir -p "$HOME/tmp" && export TMPDIR="$HOME/tmp"
cd kernel_native
make run
```

`make run` = ELF32 multiboot (QEMU `-kernel`).  
`make run64` = long mode vía ISO/GRUB.

## Controles

| Entrada | Acción |
|---------|--------|
| Ratón | Foco, arrastre, cerrar (rojo), minimizar (ámbar) |
| Dock flotante | `+` menú · W ventana · T texto · L lista · I inspector |
| **M** / **1–4** | Menú y crear organismos |
| **Enter** | Pulso Desktop → foco |
| **Tab** | Siguiente ventana |

## Manifiesto

No hay procesos de escritorio ajenos: cada superficie es un **OCB** con capacidades y pulsos. La belleza y la fluidez se construyen **dentro** de esa regla, no por encima de ella.
