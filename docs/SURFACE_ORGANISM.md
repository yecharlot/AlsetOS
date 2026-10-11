# Surface — framebuffer a la manera Alset

## Problema

Un framebuffer VGA/Bochs bare-metal es honesto con el hardware, pero limitado
para un gestor 2026 (tipografía del host, composición, multiplataforma).

## Solución Alset

**Surface** es un organismo (`kind: surface`) que representa la *superficie de presentación*.

```
Maestro
  └── Desktop     (organismo shell: iconos, menú, sesión)
        ├── Surface     (dónde se presenta el campo)
        └── Compositor  (cómo se apilan ventanas/orges)
```

| Backend | Dónde | Uso |
|---------|-------|-----|
| `html-compositor` | `alset-desktop-bridge` + `desktop/shell` | Moderno, multiplataforma, organismo en registro |
| `vbe-lfb` | `kernel_native` | Bare-metal, QEMU, sin SO anfitrión |

Ambos respetan: **el escritorio es un organismo**, no un proceso de otro SO.

## API de inspección

Con `make desktop` levantado:

```bash
curl -s http://127.0.0.1:7420/v1/organisms | jq .
```

Debes ver `sys:desktop`, `sys:surface`, `sys:compositor`.
