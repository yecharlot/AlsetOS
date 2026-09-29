# AlsetOS — Estado del repositorio

**Toolchain:** Go **1.26.0** · suite **verde** · CLIs compilables  
**Contrato:** v0.2 (`Contrato.md`)

## Qué está operativo

| Área | Estado |
|------|--------|
| Organismo + RootCID | OK |
| LispAI | OK + tests |
| Mind + Zyrion (ternario) | OK + tests |
| Genes / agentes / motor | OK (pruebas + demo) |
| Eventos / pulso | OK |
| Kernel | OK + tests |
| P2P libp2p / DHT / persistencia | OK + tests |
| Recuperación / autonomía | OK + tests |
| WASM (wazero) | OK + tests |
| Sandbox simulación | OK |
| CLIs `alsetos`, `alset-node`, `alset-kernel`, `alset-p2p` | build OK |

## Comandos

```bash
export GOTOOLCHAIN=go1.26.0
go test ./... -count=1
go run ./cmd/alsetos ejemplos/demo.alset
```

## Demo verificada

```
[ORGANISMO] organismo-demo-gtmo
[ZYRION] estado=si
[MIND] decisión=ejecutar_gene
[GENE] Gene gene-saludo ejecutado
[RESULTADO] ejecutar_gene
```

## Identidades (no mezclar)

```
RootCID  = definición del organismo
NodeID   = identidad del nodo
PeerID   = transporte libp2p
```

## Relación con PrismaTec-Core

AlsetOS = tejido experimental (manifiesto, LispAI, P2P nativo).  
PrismaTec-Core = runtime de producto (AIP, Studio, réplica TCP).  
No fusionar semánticas ternarias entre ambos.

## Pendiente

- Empaquetado operador (systemd / contenedor)
- Firma de Pulse homogénea en demos wire
- Más golden tests de RootCID
