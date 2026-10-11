# kernel_native — Desktop Alset (clon del shell)

Identidad tomada de `desktop/shell/shell.css` + comportamiento de `shell.js`,
implementado como organismos en framebuffer.

## Probar

```bash
mkdir -p "$HOME/tmp" && export TMPDIR="$HOME/tmp"
cd kernel_native && make run
```

## Guía de aprendizaje

Lee **[docs/LEARN_DESKTOP.md](../docs/LEARN_DESKTOP.md)** — mapa shell→kernel y cómo modificar el SO.

## Controles

| Acción | Cómo |
|--------|------|
| Splash → escritorio | Espera o click |
| Abrir app | Doble concepto: click en icono o menú **A** |
| Menú inicio | Botón **A** en taskbar o tecla **M** |
| Crear | **1** ventana · **2** texto · **3** lista · **4** inspector |
| Cerrar / minimizar | Botones en titlebar |
| Arrastrar | Barra de título |
| Taskbar | Click en la tarea para enfocar |

## Archivo principal

`src/kernel.c` — un solo archivo a propósito, para que puedas leerlo de arriba abajo.
