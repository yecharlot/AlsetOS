# LISPAI-AOS alineado a ATS-001

## Formas soportadas

| Forma | ATS-001 | Notas |
|-------|---------|--------|
| `(flow …)` | App / ensamblaje | Contenedor |
| `(orge …)` | Organismo | OID, kind, caps, goals |
| `(provides\|requires)` | Contratos ORGES | LEGO |
| `(bind c p contrato ver)` | Binding explícito | Uniones |
| `(state …)` | AlsetState inicial | |
| `(rule … when/then)` | ARL estructurado | Zyrion via EvalRule |
| `(ui …)` | UI-orge / IR | UiRuntime-compatible |
| `(assemble …)` | LoadAssembly | |
| `(surface …)` | Superficie activa | |
| `(test …)` | Verificación | RunFlowTests |
| `(deploy … devices …)` | Empaque multiplataforma | mobile/tablet/desktop |
| `(gene\|mind\|pulse …)` | Meta ATS | Extensión registrada |

## Studio

```bash
go run ./cmd/alset-plataforma studio --addr 127.0.0.1:5178
# http://127.0.0.1:5178/
```

Drag LEGO → edita `.aos` → Aplicar / Probar / Desplegar.
Preview: mobile · tablet · desktop.

## CLI deploy

```bash
go run ./cmd/alset-plataforma flow ejemplos/flows/mostrador.aos --deploy ./dist/cafeteria
```
