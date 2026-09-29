# Alset Desktop sobre Tiny Core Linux

**Estrategia:** no construir una distro grande y luego adelgazarla.  
Arrancar en **mínimo** (Tiny Core) y hacer de **Alset** la experiencia principal.

## Arquitectura

```
ALSET DESKTOP
├── Alset Shell          escritorio · panel · lanzador · ventanas lógicas
├── Alset-JS-Runtime     render / reactividad (en contenedor web ligero)
├── Alset Desktop Bridge IPC · procesos · archivos · eventos (Go)
├── AlsetOS Runtime      organismos · Mind · Zyrion · LispAI · Pulse
└── Tiny Core Linux      kernel · BusyBox · X · FLWM · TCZ
```

## Decisión Alset-JS (evaluación)

Alset-JS-Runtime actual dibuja sobre **DOM del navegador**, no sobre Xorg nativo.

**Prototipo 0.1:**
1. FLWM como WM de rescate / base.
2. Bridge Go local (`alset-desktop-bridge`).
3. Shell en HTML/JS servida en `localhost` y abierta con navegador mínimo o webview cuando exista.
4. AlsetOS como proceso/servicio independiente.

No se asume que JS pinte directamente en X sin contenedor.

## Orden de trabajo

1. **Tiny Core primero** — imagen arrancable, extensiones, persistencia `tce=`.
2. **Alset Shell** — UI de escritorio + bridge.
3. **Organismos** — un organismo local visible en el panel.

## Arranque y recuperación

- Si la shell Alset falla → terminal o FLWM de rescate.
- Persistencia vía mecanismos Tiny Core (`tce=`, backup), no solo tmpfs efímero.

## Hito 0.1 — criterios de aceptación

Ver `desktop/ACCEPTANCE_0.1.md`.
