# Alset Desktop + Studio + Editor

## Probar ya (desarrollo)

Desde el monorepo:

```bash
cd AlsetOS
./scripts/run-alset-desktop.sh
```

Abre:

| URL | Qué es |
|-----|--------|
| http://127.0.0.1:7420/ | **Escritorio** (iconos, menú, terminal) |
| http://127.0.0.1:7420/tools/ | **Alset Studio** |
| http://127.0.0.1:7420/tools/alset-editor/ | **Alset-JS Editor** |

Requiere la carpeta hermana `Alset-LISPAI-Runtime/web`.

## Terminal LispAI / alsetState

En el escritorio → icono **Terminal**:

```text
help
(recordar saludo hola)
(leer saludo)
(set-state count 1)
(get-state count)
status
organism
studio
editor
```

## ISO / USB

```bash
./scripts/package-alset-desktop-iso.sh
# → dist/desktop/alset-os-desktop-payload.tgz
```

En Tiny Core: extraer a `/`, ejecutar `/opt/alset/boot-alset.sh`.

## Editor

Hot-reload al editar código (App). Paneles flotantes redimensionables. Deploy PWA vía Studio API si el Studio corre en el mismo host de tools.
