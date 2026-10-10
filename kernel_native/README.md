# AlsetOS Kernel nativo (x86_64)

Todo es organismo. Sin SO anfitrión.

```
Maestro → Framebuffer → Compositor → Desktop → Ventana / Texto / Lista
```

## Requisitos

```bash
sudo apt install -y gcc qemu-system-x86
# si /tmp es solo lectura:
mkdir -p "$HOME/tmp" && export TMPDIR="$HOME/tmp"
```

## Compilar y ejecutar

```bash
cd kernel_native
make clean && make
make run
```

Usa `qemu-system-x86_64` con `-vga std`.

## Controles

| Entrada | Acción |
|---------|--------|
| Ratón | Foco / arrastre (barra de título) |
| **M** | Menú Desktop |
| **1** | Nueva ventana ORGES |
| **2** | Texto editable |
| **3** | Lista |
| **Enter** | Pulso Desktop → ventana |
| **Tab** | Siguiente ventana |
| Teclado | Escribir en texto |
| ↑↓ | Selección en lista |

## Arquitectura

- Entrada Multiboot 32-bit → long mode 64-bit
- Identity map 1 GiB (páginas 2 MiB)
- Fuente bitmap 8×8 legible
- Ratón PS/2 con suavizado y resync
