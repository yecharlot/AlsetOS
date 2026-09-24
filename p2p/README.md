# P2P de AlsetOS

## Identidades

RootCID = identidad semántica del organismo.
NodeID = identidad Ed25519 persistente del nodo.
PeerID = identidad libp2p derivada de la misma clave.
Pulse = unidad de comunicación semántica.

## Flujo

Nodo AlsetOS → identidad Ed25519 → libp2p Host → mDNS → Peer remoto → /alset/pulse/1.0.0 → Pulse firmado → verificación.

El transporte no conoce la semántica de LispAI, Mind o Zyrion. Entrega un Pulse autenticado al núcleo.

La DHT será la siguiente capa, después de validar este circuito local.
