# AlsetOS

AlsetOS es un núcleo experimental de sistema operativo orientado a **organismos digitales persistentes, verificables y recuperables**.

La ruptura conceptual es:

`programa → proceso → ejecución`

frente a:

`descripción → RootCID → organismo → Mind/LispAI/Zyrion → capacidades → Genes/Agentes → Pulso → memoria → adaptación → recuperación`

## Estado de la arquitectura

```
HARDWARE
   ↓
TinyCore / Linux
   ↓
ALSET NODE
   ├── identidad Ed25519 / NodeID
   ├── capacidades
   ├── memoria
   ├── eventos
   ├── planificador
   └── recursos
        ↓
ALSET RUNTIME
   ├── RootCID
   ├── LispAI
   ├── Zyrion
   ├── Mind
   ├── Genes
   ├── Agentes
   ├── WASM sandbox
   └── Pulse
        ↓
LIBP2P
   ├── mDNS
   ├── Pulse firmado
   ├── Kademlia DHT persistente
   ├── placement
   ├── réplicas
   └── recovery
```

## Qué significa organismo

Un organismo no es solamente un proceso.

Posee:

- identidad semántica;
- RootCID;
- memoria;
- capacidades;
- recursos autorizados;
- Genes;
- Agentes;
- lógica LispAI;
- evaluación Zyrion;
- decisiones Mind;
- eventos;
- comunicación Pulse;
- posibilidad de migración/reinstanciación.

Ciclo:

`crear → identificar → verificar → desplegar → ejecutar → recordar → adaptar → replicar → recuperar`

## Identidades

Hay tres identidades diferentes:

```
RootCID  = identidad de la definición del organismo
NodeID   = identidad criptográfica persistente del huésped
PeerID   = identidad de transporte libp2p
```

RootCID se obtiene de la representación canónica del manifiesto. Modificar la definición produce otra identidad.

## LispAI + Mind + Zyrion

Son partes del núcleo semántico, no servicios externos.

- **LispAI** interpreta instrucciones dinámicas y modifica memoria.
- **Zyrion** mantiene la lógica ternaria: no / sí / incierto.
- **Mind** transforma estado y contexto en decisiones.
- La incertidumbre no se convierte silenciosamente en falso: puede detenerse o entrar en una evaluación explícita.
- Las capacidades siguen siendo la frontera de autorización.

## Capacidades y recursos

Un organismo solamente puede utilizar una operación declarada.

Ejemplos:

```
gene.ejecutar
agente.ejecutar
wasm.ejecutar
recurso:filesystem
recurso:audio
recurso:video
```

La autorización es explícita y separada de la decisión de Mind.

## Genes y WASM

Los Genes Go son el núcleo inicial.

AlsetOS también incorpora ejecución WebAssembly mediante wazero. Un módulo WASM se ejecuta sin acceso implícito al sistema anfitrión y solamente puede ser invocado cuando el organismo posee:

`wasm.ejecutar`

Un manifiesto puede declarar:

```json
"wasm": [
  {
    "ruta": "genes/mi_gene.wasm",
    "funcion": "alset_main"
  }
]
```

La intención es convertir WASM en la unidad portable de código aislado del futuro AlsetOS.

## Memoria y eventos

La memoria conserva el estado del organismo.

El registro de eventos conserva la historia operacional mediante una cadena de hashes:

```
evento N
   ↓ hash
evento N+1
   ↓ hash
evento N+2
```

Una modificación histórica rompe la cadena y es detectable al volver a abrir el registro.

## P2P + DHT

La red ya utiliza libp2p.

Protocolos:

```
/alset/pulse/1.0.0
/alset/organism/1.0.0
```

mDNS permite descubrimiento local.

Kademlia permite localizar un organismo mediante su RootCID.

El datastore de la DHT puede persistir en Pebble para que el nodo conserve información de routing entre reinicios.

## Placement y recuperación

Un organismo puede tener:

```
RootCID
 ├── primario
 └── réplicas
```

El servicio de autonomía:

1. publica el organismo;
2. crea réplicas;
3. mantiene heartbeat;
4. detecta nodos perdidos;
5. localiza el RootCID;
6. recupera la definición;
7. verifica su integridad;
8. reinstancia el organismo;
9. actualiza el placement.

La definición nunca se acepta solamente porque otro nodo la entregue: su contenido debe producir exactamente el RootCID solicitado.

## Planificador

El planificador permite que el runtime mantenga tareas periódicas sin convertirlas en procesos permanentes independientes del organismo.

Esto prepara el modelo:

`evento → decisión → acción → nuevo estado`

## Ejecución

Núcleo local:

```bash
go run ./cmd/alsetos organismos/organismo-si.alset
```

Nodo:

```bash
go run ./cmd/alset-node organismos/organismo-persistente.alset
```

P2P:

```bash
go run ./cmd/alset-p2p
```

Autonomía:

```bash
go run ./cmd/alset-autonomia --duracion 30s
```

Publicar y replicar:

```bash
go run ./cmd/alset-autonomia \
  --publicar organismos/organismo-persistente.alset \
  --replicas 2
```

## Principio de arquitectura

El transporte no conoce la semántica de LispAI, Mind o Zyrion.

La frontera es:

```
semántica
   ↓
organismo
   ↓
Pulse
   ↓
Nodo
   ↓
transporte
```

Esto permite sustituir TCP, libp2p, WebSocket o futuros transportes sin redefinir qué es un organismo.

## Punto de llegada del runtime\n\nEl ejecutable `alset-kernel` ya reúne identidad, nodo libp2p, DHT persistente y autonomía en un ciclo de vida único. El siguiente trabajo es portar esta capa al medio de arranque TinyCore y conectar adaptadores reales de dispositivos.\n\n## Próximo salto físico

El runtime ya tiene las piezas semánticas y distribuidas necesarias para comenzar la capa de sistema operativo:

```
Alset Runtime
     ↓
capabilities
     ↓
device/resource adapters
     ↓
TinyCore Linux
     ↓
boot directo
     ↓
hardware
```

La meta no es construir otro Linux. Linux/TinyCore es el sustrato. AlsetOS debe convertirse en la capa que redefine la unidad fundamental de ejecución.
