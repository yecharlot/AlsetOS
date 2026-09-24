# Contrato del Núcleo de AlsetOS v0.2

## Propósito

AlsetOS no define una aplicación como un proceso efímero. Define una unidad semántica persistente:

`descripción → RootCID → organismo → estado → decisión → acción → memoria → adaptación`

TinyCore/Linux es el sustrato. AlsetOS es la capa que redefine la unidad de ejecución.

## Núcleo obligatorio

Los siguientes componentes forman parte del contrato:

- RootCID
- Organismo
- LispAI
- Mind
- Zyrion
- Gene
- Agente
- Memoria
- Pulso
- Capacidad
- Eventos
- Nodo
- Red P2P
- DHT
- Recovery

WASM es la frontera de aislamiento y portabilidad de código.

## Identidad

### RootCID

RootCID identifica la definición canónica de un organismo.

No identifica:

- un proceso;
- una máquina;
- una conexión;
- una instancia temporal.

### NodeID

NodeID identifica criptográficamente al huésped mediante Ed25519.

### PeerID

PeerID identifica el transporte libp2p.

Estas identidades nunca deben confundirse.

## Modelo de ejecución

```
MANIFIESTO
    ↓
CANONICALIZACIÓN
    ↓
ROOTCID
    ↓
ORGANISMO
    ↓
LISPAI
    ↓
ZYRION
    ↓
MIND
    ↓
CAPACIDADES
    ↓
GENES / WASM / AGENTES
    ↓
PULSO
    ↓
MEMORIA + EVENTOS
```

Mind decide.

Capacidad autoriza.

Gene ejecuta.

Pulso comunica.

Memoria conserva estado.

Eventos conservan historia.

## LispAI

LispAI es lenguaje interno del sistema.

Debe poder evolucionar hacia:

- evaluación simbólica;
- composición;
- introspección;
- modificación controlada del comportamiento;
- acceso a memoria;
- activación de acciones autorizadas;
- integración con Mind y Zyrion.

La ejecución nunca debe saltarse la frontera de capacidades.

## Zyrion

Zyrion conserva tres valores nativos:

```
0 = no
1 = sí
2 = incierto
```

El estado incierto es información útil y no puede degradarse silenciosamente a falso.

Mind puede:

- detener;
- evaluar;
- ejecutar una acción autorizada.

## Mind

Mind es la capa de decisión.

No posee permisos implícitos.

Una decisión de Mind nunca equivale por sí sola a autorización de recursos.

## Capacidades

Toda operación sensible requiere una capacidad explícita.

Ejemplos:

```
gene.ejecutar
agente.ejecutar
wasm.ejecutar
recurso:filesystem
recurso:audio
recurso:video
recurso:red
```

El modelo de seguridad debe ser de mínimo privilegio.

## WASM

Los módulos WASM son código potencialmente no confiable.

Por contrato:

1. se ejecutan dentro del runtime WASM;
2. no reciben acceso implícito al sistema anfitrión;
3. requieren `wasm.ejecutar`;
4. tienen límite temporal;
5. se identifican mediante su referencia dentro del manifiesto;
6. sus resultados pueden registrarse como eventos.

## Memoria

La memoria de organismo es distinta de la memoria del sistema.

Debe poder contener:

- estado;
- contexto;
- conocimiento;
- resultados;
- referencias;
- información de recuperación.

El formato inicial puede ser simple, pero su identidad pertenece al organismo.

## Eventos

Los eventos forman un historial verificable mediante encadenamiento criptográfico.

Un evento contiene:

- secuencia;
- fecha;
- tipo;
- RootCID;
- origen;
- contenido;
- hash anterior;
- hash propio.

La modificación de un evento histórico debe ser detectable.

## P2P

El transporte no define la semántica.

La arquitectura es:

```
Organismo
   ↓
Pulso
   ↓
Nodo
   ↓
Transporte
   ↓
libp2p
   ↓
DHT
```

La red actual usa:

- mDNS;
- libp2p;
- Pulse firmado;
- Kademlia;
- protocolo de organismo;
- datastore persistente.

## Recuperación

Un organismo debe poder sobrevivir a la pérdida de su huésped cuando exista una réplica compatible.

Proceso:

```
heartbeat
    ↓
detección de pérdida
    ↓
localización RootCID
    ↓
verificación de contenido
    ↓
reinstanciación
    ↓
nuevo placement
```

Nunca se debe ejecutar contenido recuperado sin verificar primero su RootCID.

## Placement

El sistema puede mantener:

```
RootCID
 ├── primario
 └── replicas
```

El número de réplicas es una política del runtime y no forma parte de la identidad semántica del organismo.

## Planificación

Las tareas periódicas pertenecen al runtime y deben poder cancelarse mediante contexto.

La ejecución prolongada no debe obligar a crear un proceso independiente por cada comportamiento.

## Hardware

AlsetOS no sustituye inmediatamente al kernel Linux.

La ruta de despliegue es:

```
Hardware
 ↓
TinyCore/Linux
 ↓
Alset Node
 ↓
Alset Runtime
 ↓
Organismos
```

La futura capa de dispositivos debe exponer recursos mediante capacidades.

## Regla fundamental

No introducir una abstracción nueva cuando el comportamiento puede expresarse naturalmente mediante:

`RootCID + Organismo + Gene + Agente + Memoria + Pulso + Capacidad + LispAI + Mind + Zyrion`

La finalidad de AlsetOS es cambiar la unidad fundamental del sistema operativo:

```
ANTES
programa → proceso

ALSET
descripción → organismo → adaptación → supervivencia
```
