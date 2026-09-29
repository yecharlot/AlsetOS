# Perfil embedded — AlsetOS en hardware de bajos recursos

Objetivo: el mismo contrato de organismo (RootCID, Mind, Zyrion, Pulse, P2P)  
sobre Raspberry Pi, edge, móvil (via contenedor/proot), IoT y nodos ligeros.

## Principios

1. **Nodo mínimo** — identidad + memoria + motor + red (sin demos pesados).
2. **Cross-compile** — `GOOS`/`GOARCH` sin CGO en el binario núcleo.
3. **Distribuido por defecto** — libp2p cuando haya red; modo local-only si no.
4. **WASM como frontera** — genes portables sin recompilar el host.
5. **Recursos acotados** — límites de memoria/CPU declarados en manifiesto.

## Targets

| Plataforma | GOOS/GOARCH | Notas |
|------------|-------------|--------|
| PC / servidor | linux/amd64 | default |
| Raspberry Pi 4/5 | linux/arm64 | |
| Pi Zero 2 / armv7 | linux/arm | softfloat si aplica |
| Windows edge | windows/amd64 | |
| macOS | darwin/arm64 | dev |
| Android (termux/proot) | linux/arm64 | experimental |
| Watch/drone firmware | — | **fase posterior**: kernel nativo; hoy = nodo Go en Linux embebido |

## Build

```bash
./scripts/build-embedded.sh
# produce dist/alsetos-linux-arm64, dist/alset-node-linux-arm, ...
```
