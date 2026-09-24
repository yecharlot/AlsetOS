# AlsetOS

AlsetOS es un núcleo experimental de sistema operativo basado en **organismos digitales persistentes**.

## Modelo

`Descripción → RootCID → Organismo → LispAI → Zyrion → Mind → Genes → Agentes → Pulso → Memoria`

El organismo es la unidad semántica de ejecución. El nodo es el huésped físico que registra, ejecuta y puede recuperar organismos.

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
          |  /
         Pulso
```

La implementación mantiene el núcleo sin dependencias externas para poder ejecutarse sobre hardware limitado. Red P2P, WASM y recuperación distribuida se incorporarán sobre esta frontera de nodo, no dentro de los conceptos semánticos básicos.

## Ejecutar el núcleo

```bash
go test ./...
go run ./cmd/alsetos organismos/organismo-si.alset
```

## Ejecutar como nodo

```bash
go run ./cmd/alset-node organismos/organismo-persistente.alset
```

Esto crea un registro persistente en `estado/registro.json`. Puedes inspeccionarlo:

```bash
cat estado/registro.json
```

El mismo nodo puede volver a levantar un organismo desde su manifiesto:

```bash
go run ./cmd/alset-node --registro estado/registro.json organismos/organismo-persistente.alset
```

## Persistencia

La memoria del organismo y el registro del nodo son dos capas distintas:

- **Memoria**: estado interno del organismo.
- **Registro RootCID**: identidad y estado conocido por el nodo.

Esta separación será importante cuando el registro pase de almacenamiento local a DHT/Red Alset.

## Principios

- RootCID identifica el organismo por contenido.
- Zyrion expresa `no / si / incierto`.
- Mind transforma el estado en decisión.
- Capacidades limitan acciones.
- Genes aportan capacidades ejecutables.
- Agentes observan y actúan.
- Pulso transporta eventos internos.
- Memoria conserva estado.
- El nodo hospeda y recupera organismos.

## Próxima frontera

La siguiente expansión técnica conecta el nodo local con:

1. ejecución de genes aislados mediante WASM;
2. transporte de Pulse entre nodos;
3. identidad y descubrimiento P2P;
4. recuperación distribuida de organismos;
5. persistencia por RootCID fuera del nodo local.

La intención no es construir una aplicación encima de un SO convencional, sino definir una unidad de ejecución distinta: el **organismo digital**.
