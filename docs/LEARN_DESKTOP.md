# Aprender el Desktop AlsetOS (tú no eres espectador)

Este documento te dice **qué archivo mirar**, **qué concepto representa** y **dónde se implementa en el kernel nativo**, para que puedas mejorar el sistema tú mismo.

---

## 1. Hay dos capas de “Desktop” (no las mezcles)

| Capa | Dónde vive | Qué es |
|------|------------|--------|
| **Desktop moderno (referencia)** | `desktop/shell/` + `cmd/alsetos-native` | UI HTML/CSS/JS servida por un núcleo ATS-001 en Go. Es el gestor “completo” que ves en desarrollo. |
| **Desktop bare-metal** | `kernel_native/src/kernel.c` | El mismo *modelo* (organismos, ventanas, taskbar, menú, iconos) dibujado a framebuffer sin Linux ni navegador. |

El kernel **no puede ejecutar** HTML/CSS. Lo que hacemos es **clonar la identidad y el comportamiento** del shell en C + framebuffer.

---

## 2. Mapa shell → kernel (copia conceptual)

| En el shell (referencia) | Archivo | En el kernel (`kernel.c`) | Función |
|--------------------------|---------|---------------------------|---------|
| Variables de color | `desktop/shell/shell.css` `:root` | `#define C_GOLD`, `C_ACCENT`, `C_BG`… | Paleta soberana Alset |
| Fondo con radiales | `#os { background: radial-gradient…}` | `draw_wallpaper()` | Atmósfera del campo |
| Iconos de escritorio | `.desk-icon` + `shell.js` DEFAULT_ICONS | `K_ICON` + `add_icon()` + `draw_icon()` | Lanzar orges |
| Ventanas | `.win`, `.win-titlebar` | `draw_window()`, `TITLE_H=38` | Orges visibles |
| Foco / z-index | `focusWin()` en `shell.js` | `focus_set()`, `z_raise()`, `z_order[]` | Quién está delante |
| Arrastre | pointer events en titlebar | `drag_win` en `mouse_feed()` | Mover organismos |
| Taskbar | `#taskbar` | `draw_taskbar()`, `TASK_H=48` | Start + tareas + estado |
| Start menu | `#start-menu` | `draw_start_menu()`, `menu_open` | Crear organismos |
| Splash | `#splash` | `phase==0`, `draw_splash()` | Arranque |
| Cursor rápido | (GPU/navegador) | `scene[]` + `cursor_move_fast()` | No redibujar todo el FB |

---

## 3. Cómo está organizado `kernel.c` (orden de lectura)

1. **Constantes y paleta** — colores del CSS.
2. **OCB (`struct OCB`)** — el organismo: nombre, kind, caps, geometría, texto, items.
3. **Framebuffer / VBE** — cómo aparece la imagen en QEMU.
4. **Primitivas de dibujo** — `fill_bb`, `panel_r`, `draw_text` (tu “motor gráfico”).
5. **Lógica de organismos** — `ocb_add`, `pulso`, `focus_set`, `crear_orges`.
6. **Compositor** — `draw_*` + `compositor_tick` + `scene_store`.
7. **Entrada** — `mouse_feed`, teclado en `kernel_main`.
8. **Bootstrap** — `bootstrap_organisms` crea Maestro, Desktop, iconos.

Regla de oro del manifiesto: **cada ventana, icono, compositor y desktop es un OCB**, no un “proceso de SO ajeno”.

---

## 4. Cómo mejorar el sistema tú mismo (rutas prácticas)

### Cambiar colores (identidad visual)
1. Abre `desktop/shell/shell.css` y mira `--gold`, `--accent`, `--bg`.
2. Copia los hex a los `#define C_*` de `kernel_native/src/kernel.c`.
3. `make run` y comprueba.

### Añadir un icono nuevo en el escritorio
1. En `bootstrap_organisms()`, llama:
   ```c
   add_icon("icNueva", "MiApp", 'N', 120, 216);
   ```
2. En `launch_from_icon()`, añade el `else if (g=='N') crear_orges(...)`.

### Añadir un tipo de ventana
1. Define un `K_NUEVO` junto a `K_TEXT`.
2. En `draw_window()`, añade un `else if (o->kind==K_NUEVO) { ... }`.
3. Expón el alta en el menú inicio (`draw_start_menu` + click en `mouse_feed`).

### Hacer el ratón aún más rápido
- El cuello de botella es **redibujar escena**.
- Hoy: movimiento solo → `cursor_move_fast`.
- Siguiente nivel: dirty-rect también al **arrastrar** (solo repintar la ventana movida).

### Entender el Desktop “completo” (HTML)
```bash
cd AlsetOS
go run ./cmd/alsetos-native
# abre http://127.0.0.1:7700/
```
Ahí editas `desktop/shell/shell.css` y `shell.js` y ves el resultado al instante. Ese es el laboratorio de UX; el kernel es el laboratorio de **organismos en metal**.

---

## 5. Honestidad técnica (para no autoengañarte)

| Expectativa | Realidad |
|-------------|----------|
| Clon pixel-perfect de GTA 6 / WebGPU | Imposible en un framebuffer 2D freestanding sin GPU |
| Clon del **comportamiento e identidad** de Alset Desktop | Esto es lo que implementa `kernel_native` |
| Tipografía Segoe UI, blur, sombras CSS | Aproximadas con bitmap 2×, paneles redondeados y blend |

El salto soberano no es copiar Windows: es que **el mismo contrato de organismos** gobierne el shell web y el kernel.

---

## 6. Comandos diarios

```bash
cd kernel_native
mkdir -p "$HOME/tmp" && export TMPDIR="$HOME/tmp"
make run          # QEMU, ELF32 multiboot
# edita src/kernel.c → make run de nuevo
```

Cuando cambies algo, anota en un comentario `/* LEARN: ... */` encima de la función. Ese hábito te convierte en co-autor del SO, no en espectador.

---

## 7. Dos caminos de arranque (por qué a veces “no ves” los cambios)

### A) Desktop moderno (recomendado para uso diario)

```bash
cd AlsetOS
git pull origin main
make desktop
# abre http://127.0.0.1:7420/
```

Aquí el **Desktop**, **Surface** y **Compositor** son organismos (`sys:desktop`, `sys:surface`, `sys:compositor`).
Surface es el “framebuffer a la Alset”: no es VGA bare-metal; es la superficie de presentación soberana del shell.

### B) Bare-metal (`kernel_native`)

```bash
cd AlsetOS
git pull origin main
cd kernel_native
make clean && make run
```

Si **no** haces `git pull` + `make clean`, QEMU seguirá el binario viejo → “sigues viendo Super Mario”.

En el splash del kernel actual debe leerse **ALSETOS 2026**. Si no aparece, no estás ejecutando el commit nuevo.
