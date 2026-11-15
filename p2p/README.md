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

El nodo dispone ahora de dos persistencias separadas:

- **Bloques de contenido:** almacén local inmutable, direccionado por el digest SHA-256 del RootCID. El CID construido por `CIDRoot` es CIDv1 con codec `raw` y multihash SHA2-256, compatible con la identificación de bloques raw de IPFS. Cada escritura y lectura vuelve a verificar el RootCID; los bloques se escriben mediante temporal, fsync y rename. La caché obtenida de otro peer también se conserva localmente.
- **Índice de organismos:** `organismos.json` conserva el índice RootCID → bytes y se publica atómicamente. Al iniciar, cada entrada se verifica contra el almacén de bloques; una entrada corrupta impide anunciarla.
- **DHT:** `ActivarPersistenciaDHT` permite utilizar Pebble como datastore persistente de Kademlia. El ejecutable `cmd/alset-p2p` lo activa por defecto en `estado/dht`; la ruta puede cambiarse con `-dht`.

La persistencia del datastore no convierte los anuncios de proveedor en disponibilidad permanente: los registros Kademlia tienen caducidad y deben renovarse. Después de conectar con un peer de bootstrap, el nodo reanuncia solo organismos cuyo bloque local supera la verificación de integridad.

`DisponibleLocal(rootCID)` informa únicamente si el bloque está localmente presente y verificado. No afirma que un proveedor remoto esté activo ni que exista disponibilidad global.

**Límite de interoperabilidad:** este cambio implementa almacenamiento local de bloques raw con CID compatible y localización de proveedores mediante Kademlia. No incorpora todavía un nodo IPFS completo, Bitswap, GossipSub, consenso, replicación automática por política ni garantía de retención entre peers.

El CID de localización no vuelve a hashear el manifiesto: se construye a partir del mismo digest SHA-256 contenido en el `RootCID`. Así, la clave Kademlia representa directamente la identidad de contenido del organismo.


## Replicación versionada

El protocolo `/alset/version/1.0.0` replica un bloque junto con un `DescriptorVersion` firmado por la identidad Ed25519 del autor. El descriptor vincula OrganismoID lógico, revisión, RootCID, antecesor, NodeID, PeerID, clave pública y fecha. La firma acredita autoría criptográfica, no autorización.

`ReplicarVersionAConocidos` realiza fan-out explícito a peers ya conectados. El índice durable `versiones.json` conserva revisiones y clasifica el linaje como `coherente`, `incompleto` o `conflicto`. Las bifurcaciones se conservan; no hay last-write-wins.

La fase siguiente incorpora GossipSub como canal de anuncios de descriptores. El protocolo de versiones continúa siendo el canal de transferencia explícita de bloques y reconciliación de contenido.


## Anuncios de versiones con GossipSub

El nodo se suscribe al topic `alset/versions/1.0.0` y publica eventos `alset.version-announcement.v1`. Cada evento transporta únicamente el descriptor firmado; no transporta los bytes del bloque.

GossipSub aplica firma y verificación estricta del mensaje libp2p. El validador de topic comprueba además el schema, el tamaño, la estructura y la firma Ed25519 del descriptor. Los eventos válidos se registran idempotentemente en el índice de versiones. Las revisiones incompletas y las bifurcaciones mantienen sus estados; el orden de llegada no selecciona ganadores.

**Separación de responsabilidades:**
- GossipSub propaga anuncios y pistas de descubrimiento.
- `/alset/version/1.0.0` transfiere el descriptor junto con el contenido cuando se solicita replicación explícita.
- RootCID permite recuperar/verificar el bloque; recibir un anuncio no prueba que el bloque esté disponible.
- La firma acredita procedencia, no autorización ni confianza.

GossipSub no conserva mensajes para nodos desconectados. La sincronización histórica de descriptores, la recuperación automática de bloques, las políticas de retención, la reparación y el factor de redundancia siguen pendientes.
