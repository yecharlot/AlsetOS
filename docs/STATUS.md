# AlsetOS — Estado

**Toolchain:** Go 1.26 · suite (verificar en entorno con toolchain 1.26)  
**Desktop host:** estrategia **Tiny Core first** (ver `docs/ALSET_DESKTOP.es.md`)  
**Génesis metal:** `kernel_native` (ver `docs/FASE0_AUDITORIA.md`)

## LISPAI-AOS (flows)

Programa ORGES en texto: `(flow … (orge …) (ui …) (assemble …))`.  
`alset-plataforma flow ejemplos/flows/mostrador.aos` — la UI es organismo; genera JSON/Go editables.  
Ver [LISPAI_AOS.md](LISPAI_AOS.md).

## Runtime organismos

Manifiesto → RootCID → Mind → gene → pulso · P2P · embedded multi-arch.

**Ciclo de vida formal (Go):** `organismo.Preparar` / `Iniciar` / `Detener` / `Fallar` — ver `organismo/ciclo.go`.

**Auditoría fundacional:** [docs/FASE0_AUDITORIA.md](FASE0_AUDITORIA.md) (matriz Manifiesto ↔ código, brechas P0).

## Alset Desktop 0.1 (en curso)

| Pieza | Estado |
|-------|--------|
| Arquitectura TC + shell + bridge | documentada |
| `alset-desktop-bridge` | código |
| Shell estática (panel/menú/organismo) | código |
| Scripts onboot / recovery | ejemplo |
| ISO USB arrancable | pendiente (en hardware/VM con TC) |

```bash
./scripts/desktop-dev.sh
# http://127.0.0.1:7420/
```

Orden: **Tiny Core → Alset Shell → organismos**.
