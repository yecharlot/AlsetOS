# Guía AlsetOS — Escritorio, Kernel y ORGES (descentralizado-first)

## Qué es esto (sin humo)

AlsetOS no pretende ser “otro Linux”. La unidad fundamental es el **organismo** (OCB: Organism Control Block), no el proceso.

Hay **dos superficies** del mismo modelo:

1. **Kernel nativo** (`kernel_native/`) — arranca en QEMU **sin SO anfitrión**. Escritorio en VGA: lista de organismos, Pulse local, creación de ORGES/Peer.
2. **Alset Desktop / ORGES Studio** (`cmd/alsetos-native`) — el escritorio rico (ventanas, Studio no-code, export PWA) que ya conoces. Hoy corre como binario Go en tu PC para **crear y desplegar**; comparte la misma filosofía (Master, Pulse, ORGES, capacidades).

No es que el HTML del Desktop se ejecute dentro del kernel de texto VGA (haría falta un motor de navegador en el núcleo). Es que **Desktop, Studio, Master y ORGES son organismos** en ambos lados, y el kernel ya los instancia al arrancar.

## Principio descentralizado-first

- Cada nodo tiene un **NodeID** local (no hay registro central).
- **Master** es soberanía **del nodo**, no un admin global de Internet.
- Toda interacción entre organismos es **Pulse** (mensaje con nonce), no APIs centrales obligatorias.
- Export PWA: la app generada puede vivir offline (service worker); el origen de verdad del organismo sigue siendo el nodo que lo creó / su RootCID cuando uses el runtime Go.
- Peers se añaden como organismos `PEER` en el kernel; la malla P2P completa (libp2p) se engancha al mismo modelo de identidad, no al revés.

## Arranque rápido

### A) Kernel (sin Linux anfitrión dentro de la VM)

```bash
cd kernel_native
make
make run
```

Controles en QEMU:

| Tecla | Acción |
|-------|--------|
| ↑ / ↓ | Seleccionar organismo |
| Enter | Pulse Master → seleccionado |
| 1 | Crear organismo ORGES |
| 2 | Crear organismo Peer (descentralizado) |
| 3 | Pulse Desktop → Studio (tick Studio) |
| 4 | Broadcast Pulse desde Desktop |
| Esc | Redibujar |

Cerrar la ventana de QEMU para salir.

### B) Escritorio rico + ORGES Studio (desarrollo y export)

```bash
cd /ruta/a/AlsetOS
go build -o bin/alsetos-native ./cmd/alsetos-native
./bin/alsetos-native -addr 127.0.0.1:7700
```

Abre http://127.0.0.1:7700/

- Usuario: `Master` / contraseña: `master`
- Icono **ORGES Studio**: no-code, código, preview Web/Escritorio/Android
- **Desplegar organismo** → queda en el nodo Go
- **Exportar Web/PWA** → ZIP instalable (manifest + service worker)

## Mapa mental (resultado final deseado)

```
                    ┌─────────────────────────────┐
                    │   Nodo Alset (soberano)     │
                    │   NodeID · Master OCB       │
                    └─────────────┬───────────────┘
                                  │ Pulse
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
   Desktop OCB              Studio OCB                  ORGES / Peer
   (shell / UI)           (crear / export)            (dominio negocio)
        │                         │                         │
        └─────────────────────────┴─────────────────────────┘
                                  │
                    malla de peers (descentralizado)
```

- **Ahora en kernel:** Desktop, Studio, Master, PulseBus como OCB + UI VGA interactiva.
- **Ahora en Go:** misma semántica con UI completa y export PWA.
- **Dirección:** ir moviendo capacidades del Studio al nodo (y más tarde drivers/red en kernel) sin introducir un servidor central.

## Buenas prácticas al crear ORGES

1. Define **goals** y **capabilities** antes que la UI.
2. Todo acceso entre organismos = Pulse (no “atajos” de memoria compartida ad hoc).
3. Exporta PWA para distribución; el organismo en el nodo sigue siendo la autoridad de identidad.
4. Añade Peers solo con las caps mínimas (least privilege).
5. Piensa multi-nodo desde el día 1: nada que solo funcione con “localhost mágico” como autoridad.

## Límites actuales (realismo)

- El kernel fase integrada **no** renderiza el HTML del Studio; muestra el **modelo Desktop** en VGA.
- Red P2P real (libp2p) y persistencia durable en kernel: siguientes incrementos sobre el mismo OCB/Pulse.
- APK Android nativo de tienda: el export actual es **PWA**; Capacitor/Cordova sería empaquetado posterior del mismo ZIP web.

## Un solo repositorio

```
AlsetOS/
  kernel_native/     → kernel multiboot + desktop OCB (QEMU)
  cmd/alsetos-native → escritorio rico + ORGES Studio + PWA
  internal/ats001/   → Zyrion, Pulse, organism, kernel lógico Go
  docs/GUIA_ALSETOS.md → esta guía
```

Trabaja siempre con la pregunta: *¿esto es un organismo con OID/caps/Pulse, o estoy reintroduciendo un proceso/servidor central?*
