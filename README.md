# AlsetOS

AlsetOS es un núcleo experimental de sistema operativo basado en **organismos digitales persistentes**.

## Modelo

`Descripción → RootCID → Organismo → LispAI → Zyrion → Mind → Genes → Agentes → Pulso → Memoria`

El **organismo** es la unidad semántica de ejecución. El **nodo** es el huésped físico que registra, ejecuta, recupera y conecta organismos.

## Arquitectura actual

```
Hardware / TinyCore-Linux
        ↓
    alset-node
        ↓
 RootCID Registry
        ↓
    Organismo
   ↙    ↓     ↘
LispAI Zyrion  Mind
   ↓      ↓      ↓
Memoria Genes  Agentes
        \  |  /
         Pulso
           ↓
     transporte firmado
           ↓
       Nodo remoto
```

La implementación mantiene el núcleo sin dependencias externas para poder ejecutarse sobre hardware limitado.

## Identidad del organismo

El RootCID se calcula sobre la **definición canónica completa del manifiesto**:

```
manifiesto
    ↓
representación canónica
    ↓
SHA-256
    ↓
RootCID
```

Así, la identidad pertenece a la definición del organismo y no al proceso que lo ejecuta.

## Identidad del nodo

Cada `alset-node` posee una identidad criptográfica Ed25519 persistente:

```
clave privada + clave pública
            ↓
         NodeID
```

La identidad se crea automáticamente en:

```
estado/identidad.json
```

La clave privada queda excluida de Git mediante `.gitignore`.

Para ejecutar:

```bash
go run ./cmd/alset-node organismos/organismo-persistente.alset
```

El nodo mostrará su `NODE-ID`. Las ejecuciones posteriores reutilizan la misma identidad.

## Pulse autenticado

La capa `red/` ya puede transportar un Pulse firmado:

```
Nodo A
  │
  ├── NodeID
  ├── clave pública
  ├── Pulse
  └── firma Ed25519
          │
          ↓
       Nodo B
          │
          ↓
       verificar
          │
          ↓
        aceptar
```

Un receptor rechaza el Pulse si el contenido fue manipulado después de firmarse.

El transporte actual usa `net.Conn`/TCP como laboratorio mínimo. No es la arquitectura P2P definitiva.

## Frontera de red

La separación deliberada es:

```
Organismo
   ↓
  Pulso
   ↓
  Nodo
   ↓
 Transporte
   ↓
 libp2p (próxima capa)
   ↓
 DHT / descubrimiento
   ↓
 Nodo remoto
```

La semántica de `Organismo`, `Mind`, `Zyrion` y `Pulso` no depende del transporte.

## Persistencia

Existen dos capas distintas:

- **Memoria**: estado interno del organismo.
- **Registro RootCID**: identidad y estado conocido por el nodo.
- **Identidad del nodo**: credencial criptográfica del huésped.

## Próxima frontera

La siguiente expansión importante es:

1. descubrimiento entre nodos;
2. protocolo de conexión nodo↔nodo;
3. adaptación del transporte a libp2p;
4. recuperación de organismos por RootCID;
5. genes aislados mediante WASM;
6. persistencia distribuida y DHT.

La intención sigue siendo definir una unidad de ejecución distinta: el **organismo digital**.
