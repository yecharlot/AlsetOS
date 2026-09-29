# AlsetOS — Status

**Toolchain:** Go 1.26 · green suite · embedded cross-compile  
**Contract:** v0.2

AlsetOS is the **organism OS layer**: RootCID, Mind, Zyrion, Pulse, P2P, recovery.  
Linux is substrate; the product is the multi-platform **Alset node**.

### Working today

- Manifest → RootCID → Mind → gene → pulse  
- libp2p P2P + tests  
- Embedded multi-arch builds (`scripts/build-embedded.sh`)  
- CGO-free CLIs  

### Roadmap

1. Portable Go node (Pi, PC, edge)  
2. Raspberry image / minimal container  
3. Alset-JS shell over node AIP  
4. Heterogeneous fleet placement/recovery  
5. Ultra-light profiles (no DHT) via gateway for wearables/drones  

See `profiles/embedded/README.md`.
