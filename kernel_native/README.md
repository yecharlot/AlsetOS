# AlsetOS Kernel Native

Kernel **propio** (multiboot), sin Linux ni TinyCore como anfitrión del núcleo.

## Fase 1 (actual)

- Arranque Multiboot en QEMU
- Consola VGA texto
- **Organism Control Block (OCB)** Master en memoria
- Sin procesos Unix: la unidad es el organismo

## Compilar y probar

```bash
cd kernel_native
make
make run          # requiere qemu-system-i386 o qemu-system-x86_64
# opcional:
make iso          # requiere grub-mkrescue
```

## Relación con AlsetOS Native (Go)

| Capa | Qué es | Dónde corre |
|------|--------|-------------|
| `kernel_native` | Kernel metal / QEMU | Sin SO anfitrión |
| `cmd/alsetos-native` | Runtime ATS-001 + escritorio + ORGES Studio | Hoy sobre host (Ubuntu/Windows) como puente de desarrollo |

El runtime Go **no es el kernel**. El camino es portar primitivas (OCB, Pulse, scheduler) del modelo ATS al kernel C/ensamblador. El escritorio y ORGES Studio siguen siendo el entorno de desarrollo hasta que el kernel tenga framebuffer/red suficientes.

## Honestidad

Fase 1 **no** incluye GUI, red, ni ORGES Studio dentro del kernel. Demuestra arranque soberano y OCB. Las fases siguientes: scheduler de OCBs, Pulse interno, persistencia, drivers.
