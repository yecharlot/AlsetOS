# AlsetOS Kernel nativo — Desktop por organismos

Todo es organismo. Sin SO anfitrión.

```
Maestro → Framebuffer → Compositor → Desktop → Orges (ventana / texto / lista / inspector)
```

## Salto de esta versión

| Antes | Ahora |
|-------|--------|
| Redibujado crudo al FB | **Doble buffer** + present |
| Fuente 8×8 ilegible | Fuente **2× (16 px)** legible |
| Ratón limitado / errático | **Pantalla completa**, resync PS/2, aceleración leve |
| Sin z-order | **Pila z-order** y foco |
| UI mínima | **Top bar + dock + menú + sombras + gradiente** |
| Sin inspector | **Inspector OCB** de la red de organismos |

## Arranque (QEMU)

```bash
mkdir -p "$HOME/tmp" && export TMPDIR="$HOME/tmp"
cd kernel_native
make run
```

`make run` usa **ARCH=32** porque `qemu -kernel` solo carga ELF multiboot 32-bit.

Kernel **x86_64** long mode:

```bash
sudo apt install -y grub-pc-bin grub-common xorriso
make run64
```

## Controles

| Entrada | Acción |
|---------|--------|
| Ratón | Foco, arrastre por barra de título, cerrar (rojo), minimizar (ámbar) |
| Dock | Menu / Win / Text / List / Insp |
| **M** | Menú crear organismo |
| **1–4** | Ventana / Texto / Lista / Inspector |
| **Enter** | Pulso Desktop → ventana en foco |
| **Tab** | Siguiente ventana |
| Flechas | Mover ventana o navegar lista |
| Texto | Escribir en editor con cursor |

## Filosofía

No hay “procesos de Windows”. Cada ventana, el compositor, el framebuffer y el desktop son **organismos (OCB)** con capacidades y pulsos. El salto visual no rompe ese contrato.
