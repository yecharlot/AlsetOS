# Runtime declarativo + dominio pedidos (rebanada vertical ATS-001)

## 1. Auditoría breve (situación real)

| Capacidad ATS-001 | Estado en repo antes de esta rebanada |
|-------------------|----------------------------------------|
| Organismo + RootCID + ciclo de vida | Parcial (Go: `organismo`, `manifiesto`, `ciclo`) |
| Capacidades | Parcial (`capacidad.Requerir`) |
| Zyrion 0/1/2 | Implementado (`zyrion`) |
| Mind | Parcial (decide Si/No/Incierto; **no debe** otorgar caps) |
| Pulse firmado Ed25519 | Identidad Ed25519 existe; Pulse simple **sin** firma |
| AlsetState reactivo con deps | Solo átomos sin dependencias (`internal/ats001/state`) |
| LISPAI-AOS parser/AST | **Ausente** (LispAI mínimo: `recordar`) |
| ARL parser/gramática texto | **Ausente** (solo especificación en PDF) |
| Resolución de contratos ORGES | **Ausente** |
| Ensamblaje validado | **Ausente** |

Esta entrega **no inventa** un parser LISPAI-AOS de texto libre. Extiende el modelo con **declaraciones estructuradas (JSON/Go)** y **ARL estructurado (AST)**, alineados a los campos del ATS-001.

## 2. Flujo implementado

```
OrgDecl (JSON/Go)
  → Validate + DefinitionCID
  → Resolve (requires ↔ provides, semver, anti-ambiguity, ciclos)
  → Runtime.LoadAssembly (organismo.Preparar/Iniciar + caps)
  → ReactiveState + EvalRule (Zyrion)
  → RequireCap (autorización separada de Mind)
  → SignedPulse (Ed25519 + nonce anti-replay)
  → Persist / Restore
```

## 3. Dominio pedidos

Organismos: **Pedido**, **Inventario** (A o B), **Presupuesto**, **Aprobación**, **Auditoría**.

Prueba de valor: sustituir Inventario A → B **sin modificar** la función `Pedido()` ni sus reglas.

## 4. Cómo ejecutar

```bash
cd AlsetOS
# tests
go test ./declarativo/ ./ejemplos/pedidos/ -count=1

# demo
go run ./cmd/alset-pedidos
go run ./cmd/alset-pedidos -provider=b
```

Si `go.mod` exige 1.26 y solo tienes 1.22:

```bash
# solo para prueba local
sed -i 's/go 1.26.0/go 1.22/' go.mod   # revertir después
```

## 5. Seguridad

| Tema | Estado |
|------|--------|
| Firmas Pulse Ed25519 | Implementado en `declarativo.EmitPulse` / `VerifyPulse` |
| Anti-replay nonce | Implementado |
| Cap check en acciones | `RequireCap` / `HasCap` |
| Mind no otorga caps | Cubierto por test |
| Aislamiento de procesos | **No** afirmado (runtime en un solo proceso Go) |
| Credenciales master en rutas públicas | No introducidas en esta rebanada |

## 6. Siguiente fase (corta)

1. Parser textual ARL/LISPAI-AOS mínimo → AST ya usado aquí.
2. Señales AlsetState (push) además de dirty-set.
3. Conectar ensamblaje a `cmd/alsetos-native` HTTP sin meter lógica en `kernel_native`.
