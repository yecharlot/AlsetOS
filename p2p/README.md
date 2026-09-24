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


## DHT y recuperación por RootCID

La fase DHT añade tres operaciones:

1. `AnunciarOrganismo`: valida que el manifiesto produzca exactamente el RootCID y anuncia ese identificador mediante un registro de proveedor Kademlia.
2. `RecuperarOrganismo`: localiza proveedores del RootCID y solicita el manifiesto por `/alset/organism/1.0.0`.
3. `recuperacion.Servicio.EjecutarRemoto`: valida el manifiesto recuperado y lo entrega al Motor local para su ejecución.

La DHT no transporta la semántica de LispAI, Mind o Zyrion. Solo localiza al nodo que puede entregar la definición. La definición se verifica de nuevo calculando su RootCID antes de aceptarla.

Flujo: RootCID → CID de localización → Kademlia DHT → proveedor → `/alset/organism/1.0.0` → manifiesto → verificación RootCID → Motor → organismo ejecutado localmente.

El datastore DHT actual es de laboratorio y vive en memoria. La siguiente evolución es persistirlo junto con el registro de organismos para soportar reinicio y recuperación automática.

El CID de localización no vuelve a hashear el manifiesto: se construye a partir del mismo digest SHA-256 contenido en el `RootCID`. Así, la clave Kademlia representa directamente la identidad de contenido del organismo.
