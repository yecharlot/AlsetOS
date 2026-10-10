# AlsetOS Kernel nativo

Todo es organismo. Sin SO anfitrión.

```
Maestro → Framebuffer → Compositor → Desktop → Ventana / Texto / Lista
```

## Importante (QEMU)

`qemu -kernel` **solo carga ELF multiboot de 32 bit**.  
Por eso `make run` usa **ARCH=32** (con fuente legible y ratón estable).

El kernel **x86_64 real** se arranca con ISO/GRUB:

```bash
make run64   # genera ISO y arranca qemu-system-x86_64 -cdrom
```

## Uso diario (pantalla gráfica)

```bash
mkdir -p "$HOME/tmp" && export TMPDIR="$HOME/tmp"
cd kernel_native
make run
```

## Kernel 64-bit

```bash
sudo apt install -y grub-pc-bin grub-common xorriso
make run64
```

## Controles

| Entrada | Acción |
|---------|--------|
| Ratón | Foco / arrastre (suavizado) |
| **M** | Menú |
| **1 / 2 / 3** | Ventana / Texto / Lista |
| **Enter** | Pulso |
| **Tab** | Siguiente ventana |

## Nota sobre “64 bit”

- **HOST**: se recomienda `qemu-system-x86_64`.
- **Kernel por defecto (`make run`)**: código en modo protegido 32-bit (exigencia de multiboot `-kernel`).
- **Kernel long mode 64-bit**: `make run64` / `make iso`.
