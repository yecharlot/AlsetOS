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
     transporte de red
```

La implementación mantiene el núcleo sin dependencias externas para poder ejecutarse sobre hardware limitado.

## Identidad RootCID

El RootCID ya no depende solamente del nombre del organismo.

Ahora se calcula sobre la **definición canónica completa del manifiesto**:

```
manifiesto
    ↓
representación canónica
    ↓
SHA-256
    ↓
RootCID
```

Por tanto:

- mismo manifiesto → mismo RootCID;
- manifiesto diferente → RootCID diferente;
- cambiar capacidades, Genes, LispAI o configuración semántica cambia la identidad.

Esto establece una frontera importante: **la identidad pertenece a la definición del organismo, no al proceso que lo ejecuta**.

## Ejecutar el núcleo

```bash
go test ./...
go run ./cmd/alsetos organismos/organismo-si.alset
```

## Ejecutar como nodo

```bash
go run ./cmd/alset-node organismos/organismo-persistente.alset
cat estado/registro.json
```

El registro persistente mantiene la relación:

```
RootCID → organismo → último estado conocido
```

## Transporte Pulse

La nueva capa `red/` transporta Pulsos sin introducir semántica de red en el organismo:

```
Organismo
    ↓
  Pulso
    ↓
Nodo
    ↓
Transporte
    ↓
Nodo remoto
    ↓
  Pulso
```

La implementación actual usa TCP mediante `net.Conn` y JSON delimitado por líneas como transporte mínimo y comprobable.

**Importante:** TCP no es la arquitectura final. La capa está diseñada para que posteriormente pueda sustituirse por libp2p/DHT sin modificar la semántica de `Pulso`, `Organismo`, `Mind` o `Zyrion`.

## Persistencia

La memoria del organismo y el registro del nodo son dos capas distintas:

- **Memoria**: estado interno del organismo.
- **Registro RootCID**: identidad y estado conocido por el nodo.

Esta separación será la base para recuperación distribuida.

## Principios

- RootCID identifica la definición del organismo.
- Zyrion expresa `no / si / incierto`.
- Mind transforma el estado en decisión.
- Capacidades limitan acciones.
- Genes aportan capacidades ejecutables.
- Agentes observan y actúan.
- Pulso transporta eventos.
- Memoria conserva estado.
- El nodo hospeda, registra y recupera organismos.
- La red transporta Pulsos sin contaminar el núcleo semántico.

## Próxima frontera

La siguiente expansión importante será:

1. **identidad de nodo** y claves criptográficas;
2. **descubrimiento entre nodos**;
3. sustitución/adaptación del transporte a **libp2p**;
4. **Pulse entre organismos reales**;
5. recuperación de la definición por RootCID;
6. genes aislados mediante **WASM**;
7. después, persistencia distribuida y DHT.

La intención no es construir una aplicación encima de un SO convencional, sino definir una unidad de ejecución distinta: el **organismo digital**.
