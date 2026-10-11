# Fase 0 — Auditoría AlsetOS (Manifiesto, contratos, código real)

**Fecha:** 2026-10-11  
**Repo:** https://github.com/yecharlot/AlsetOS  
**Commit base de esta auditoría:** `044c071` (+ cambios de esta iteración)

Este documento **no inventa** el Manifiesto. Resume lo verificado en el árbol del repositorio y marca lagunas con honestidad.

---

## 1. Documentos normativos localizados

| Documento | Ubicación | Rol normativo |
|-----------|-----------|---------------|
| **Contrato del Núcleo v0.2** | `Contrato.md` | Contrato rector técnico del núcleo semántico |
| Manifiesto (código) | `manifiesto/manifiesto.go` | Estructura JSON canónica → RootCID |
| Guía Génesis | `docs/GUIA_GENESIS.md` | Modelo kernel_native (organismos-dispositivo, Pulso) |
| Organismos y sistema | `docs/ORGANISMOS_Y_SISTEMA.es.md` | Bridge desktop / APIs REST |
| STATUS | `docs/STATUS.md` | Estado declarado: Tiny Core first + suite Go |
| Desktop | `docs/ALSET_DESKTOP.es.md` | Estrategia shell + bridge |
| UiRuntime | `docs/UI_RUNTIME.es.md` | IR de apps en kernel_native |
| Learn Desktop | `docs/LEARN_DESKTOP.md` | Mapa shell ↔ kernel |

**No se encontró** un archivo separado titulado exactamente «Manifiesto.md». El contrato vigente en texto es **`Contrato.md`**. El manifiesto *de organismo* es el tipo Go `manifiesto.Manifiesto` (JSON).

### Principios del Contrato (verificados en texto)

1. Unidad: `descripción → RootCID → organismo → estado → decisión → acción → memoria → adaptación`
2. Identidades distintas: **RootCID** ≠ **NodeID** ≠ **PeerID**
3. Mind decide; **Capacidad autoriza**; Gene/WASM ejecuta; Pulso comunica; Memoria conserva
4. Zyrion ternario: 0/1/2; el 2 (incierto) no se degrada a falso
5. Mínimo privilegio; WASM sin acceso implícito al anfitrión
6. Recuperación solo tras verificar RootCID

---

## 2. Contradicción documental (no se resuelve en silencio)

| Fuente | Afirma |
|--------|--------|
| `Contrato.md` | «**TinyCore/Linux es el sustrato**. AlsetOS es la capa que redefine la unidad de ejecución.» |
| `docs/STATUS.md` | Desktop: estrategia **Tiny Core first** |
| `README.md` | Diagrama HARDWARE → TinyCore/Linux → ALSET NODE |
| Misión fundacional actual | **`kernel_native` es la base**; no sustituir por Linux como núcleo del producto final |

**Decisión mínima compatible con el Manifiesto conceptual (“todo es organismo”) y con el código existente:**

- **Track A (producción semántica hoy):** runtime Go + ATS-001 + P2P sobre Linux/TinyCore *como huésped*, no como definición de “organismo”.
- **Track B (génesis bare-metal):** `kernel_native` es el **organismo Kernel** objetivo; aún no ofrece frontera de usuario aislada.
- **No hay un único producto integrado** que unifique A y B en un solo arranque. Hay dos superficies de ejecución. La arquitectura objetivo es: B provee mecanismos; A (o un runtime portado) vive como organismos de usuario cuando exista frontera real.

Cualquier cambio a `Contrato.md` que declare solo `kernel_native` debe hacerse de forma explícita y versionada (v0.3), no por omisión en demos gráficas.

---

## 3. Matriz de trazabilidad (principio → implementación → prueba)

| Principio Contrato | Contrato / tipo | Implementación real | Pruebas | Estado |
|--------------------|-----------------|---------------------|---------|--------|
| RootCID identidad de definición | `rootcid/`, manifiesto canónico | `rootcid/rootcid.go`, `manifiesto.Canonico()` | `pruebas/rootcid_test.go`, `manifiesto_test.go` | **Implementado** (Go) |
| Organismo + ciclo de vida | `organismo.Estado` | Constantes creado/listo/ejecutando/detenido/error; **solo `Nuevo()`** | `organismo_test.go` solo crea | **Parcial** — sin transiciones formales |
| Capacidades | `capacidad/`, ATS Cap | Maps y `HasCap`; bridge y ATS-001 | `pruebas/capacidad_test.go` | **Implementado** (Go); kernel usa bitmasks |
| Mind no autoriza | Contrato § Mind | `mind/`, `internal/ats001/mind` | `pruebas/mind_*.go` | **Implementado** (lógica Go) |
| Zyrion 0/1/2 | `zyrion/` | Valores No/Si/Incierto | `zyrion_test.go` | **Implementado** |
| LispAI | `lispai/` | Evaluador | `lispai_test.go` | **Implementado** (Go) |
| Gene / WASM | `gene/`, `wasm/` | Genes + wazero | pruebas gene/wasm | **Implementado** (Go host) |
| Pulso | `pulso/`, ATS pulse | Mensajes / firmas en red | `pulso_test.go`, red | **Implementado** (Go); kernel_native pulso = contadores |
| Memoria organismo | `memoria/` | map estado | pruebas | **Parcial** |
| Eventos encadenados | `eventos/` | historial | pruebas | **Parcial** |
| P2P / DHT | `p2p/` | libp2p, DHT persistente | `p2p/*_test.go` | **Implementado** (Go/Linux) |
| Recovery | `recuperacion/` | servicio | pruebas | **Parcial** |
| ATS-001 microkernel | `internal/ats001/kernel` | Registro orgs, caps, audit, BootMaster | vía desktop-bridge / native | **Implementado en proceso Go**, no en metal |
| Kernel bare-metal | `kernel_native/` | boot, FB, OCB, UI, ratón | `make` compila; QEMU manual | **Prototipo**: sin procesos de usuario, sin syscall, sin FS, sin red real |
| Desktop moderno | `desktop/shell`, bridge | HTML/JS + HTTP | scripts dev | **Demo host**; no reiniciable independiente del kernel metal |
| UiRuntime | IR en `kernel.c` | K_APP embebido | visual QEMU | **Prototipo in-kernel** (viola meta largo plazo “UI fuera del núcleo”) |

---

## 4. Auditoría `kernel_native` (capacidades reales)

| Área | ¿Existe en código? | Evidencia |
|------|--------------------|-----------|
| Arranque multiboot | Sí | `boot/boot32.S`, `boot/boot.S` |
| Long mode 64-bit | Parcial | `boot.S` + ISO/GRUB; QEMU `-kernel` solo ELF32 |
| Framebuffer / VBE | Sí | `fb_init_*`, backbuffer, scene cache |
| Teclado / ratón PS/2 | Sí | polling puerto 0x60/0x64 |
| OCB (organismos in-kernel) | Sí | `struct OCB`, caps bitmask |
| Pulso entre OCB | Mínimo | incrementa contadores pin/pout |
| Aislamiento de privilegios (ring 3) | **No** | todo en el mismo espacio |
| Tabla de páginas por proceso | **No** | identity map solo en boot 64 |
| Syscalls | **No** | — |
| Planificador de procesos | **No** | bucle único `kernel_main` |
| Temporizador / IRQ | **No** | busy-wait `pausa` |
| Sistema de archivos | **No** | — |
| Red | **No** | — |
| Persistencia post-reinicio | **No** | estado solo RAM |
| Escritorio / UiRuntime | Sí, **dentro del kernel** | compositor + IR |

**Conclusión Fase 1:** el bloqueo real no es “hacer el desktop más bonito”. Es **ausencia de frontera kernel / usuario**. Sin ella no se puede cumplir el criterio: “un servicio falla sin derribar el núcleo” ni “el escritorio se reinicia sin reiniciar el kernel”.

---

## 5. ATS-001 y runtime Go

`internal/ats001/kernel` es un **microkernel lógico en proceso de usuario Linux**:

- BootMaster con caps soberanas  
- Create organism, pulse, audit, checkpoint a disco del host  
- No gestiona CPU, MMU ni dispositivos reales  

El paquete raíz `kernel/kernel.go` arranca **nodo P2P + autonomía**, no un kernel de hardware.

El tipo `organismo.Organismo` declara estados de ciclo de vida pero **no implementa transiciones** (solo `Nuevo` → `creado`).

---

## 6. Brechas priorizadas (orden de dependencia)

| Prioridad | Brecha | Por qué bloquea |
|-----------|--------|-----------------|
| P0 | Frontera de ejecución usuario en `kernel_native` (o decisión explícita de huésped hasta entonces) | Sin esto el desktop “completo” en metal es monolito privilegiado |
| P0 | Unificar narrativa Contrato v0.2 (TinyCore) vs génesis bare-metal | Evita dos productos incompatibles presentados como uno |
| P1 | Ciclo de vida formal del organismo (transiciones + precondiciones) | Contrato común de organismos |
| P1 | Capacidades: alinear nombres string (Go) y bitmasks (C) | Mismo vocabulario de autoridad |
| P2 | Extraer UiRuntime/compositor de lógica “app” hacia contrato de servicio | Desktop reiniciable |
| P2 | Persistencia RootCID en metal o vía puente verificable | Supervivencia post-reinicio |
| P3 | Planificador, FS, red en metal | Niveles B del manifiesto de misión |

---

## 7. Arquitectura objetivo (sin segunda filosofía)

```
Nivel A  Organismo Kernel (kernel_native)
           mecanismos: boot, memoria, IRQ, (futuro) syscall + aislamiento
Nivel B  Organismos de sistema (servicios con mínimo privilegio)
Nivel C  ATS-001 / RootCID / Mind / Zyrion / LispAI / Gene / Pulso
           hoy: proceso Go en huésped; meta: sobre frontera de A
Nivel D  Desktop = composición de organismos (shell, compositor, WM)
           hoy: shell HTML + bridge; meta: fuera del núcleo
Nivel E  Aplicaciones = organismos con caps y manifiesto
```

**Primera pieza LEGO (esta iteración):** formalizar ciclo de vida del organismo en el contrato común Go y reflejar estados de fase en OCB del kernel, con pruebas y documentación — sin fingir aislamiento de procesos.

---

## 8. Límites de esta auditoría

- Suite Go: `go.mod` exige **1.26**; el entorno de auditoría tenía **1.22** y falló la descarga del toolchain (`502`). Las pruebas Go **no se re-ejecutaron aquí** con evidencia verde.
- Compilación `kernel_native` ARCH=32: **OK** en iteración previa (`make`).
- QEMU visual no automatizado en este entorno.
