# AlsetOS — Estado

**Toolchain:** Go 1.26 · suite verde · cross-compile embedded  
**Contrato:** v0.2

## De experimental → sistema operativo de organismos

AlsetOS **deja de presentarse solo como prototipo de laboratorio**.  
El contrato (organismo, RootCID, Mind, Zyrion, Pulse, P2P, recovery) es la unidad de ejecución.  
Linux/TinyCore es el sustrato; el producto es el **nodo Alset** multiplataforma.

### Ya funcional

- Ciclo manifiesto → RootCID → Mind → gene → pulso  
- P2P libp2p + tests  
- Perfil **embedded** + script multi-arch (`scripts/build-embedded.sh`)  
- CLIs compilables sin CGO  

### Trayectoria de producto

| Fase | Entrega |
|------|---------|
| Ahora | Nodo Go portable (Pi, PC, edge) |
| +1 | Imagen Raspberry / contenedor mínimo |
| +2 | Shell/UI Alset-JS sobre AIP del nodo |
| +3 | Placement y recovery en flotas heterogéneas |
| +4 | Perfiles ultra-light (sin DHT) para wearables/drones vía gateway |

Ver `profiles/embedded/README.md`.
