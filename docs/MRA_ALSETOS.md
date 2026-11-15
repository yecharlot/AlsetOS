# MRA y su integración progresiva en AlsetOS

## Propósito
Incorporo el Modelo de Relaciones Abstraccionales (MRA) como fundamento formal para describir las relaciones que constituyen y condicionan a los organismos. MRA no es un segundo runtime ni reemplaza el modelo actual.

## Contrato de compatibilidad
1. Los manifiestos existentes siguen siendo válidos.
2. Los campos relacionales nuevos serán opcionales y aditivos.
3. RootCID seguirá identificando la definición canónica; no será identidad de instancia, nodo o relación.
4. NodeID y PeerID conservan sus responsabilidades.
5. Una relación no concede capacidades ni autorización.
6. Mind evalúa; Zyrion conserva incertidumbre; Capacidad autoriza.
7. No se infiere causalidad de una asociación.
8. No se asigna un nivel numérico universal contando relaciones.
9. No se presupone conciencia por memoria o autorrepresentación.
10. El transporte P2P no adquiere semántica de dominio.

## Contrato numérico de Zyrion
El contrato canónico de estados es:
- `1` = True (verdadero).
- `0` = False (falso).
- `2` = Indeterminado.

Este mapeo es explícito tanto en `zyrion` como en `internal/ats001/zyrion`. No se debe invertir el significado de los valores ni usar el orden de declaración de constantes para reinterpretarlos.

El estado `2` es un estado sumidero para el nivel de evaluación actual y activa la evaluación del siguiente nivel de abstracción. En ese nivel se aplican las mismas reglas declaradas, sobre el contexto del organismo relacionado. La recursividad no implica profundidad infinita: se controlan ciclos, profundidad y cantidad de organismos examinados. Si no existe una regla equivalente, faltan datos, se alcanza un límite o las ramas discrepan, el resultado permanece en `2`.

La política inicial de convergencia es deliberadamente conservadora: un nivel solo se resuelve si todos sus hijos aplicables terminan en el mismo estado determinado (`0` o `1`). Una mezcla de resultados, un hijo indeterminado o una rama cíclica no se convierte en certeza. Esta política es una decisión de implementación versionada, no una ley universal de la lógica fractal.

## Modelo
Una abstracción computacional se describe mediante identidad, estructura, relaciones, capacidades y restricciones, estado y procedencia.

Una relación contiene ID, origen, destino, tipo abierto, dirección opcional, temporalidad, condiciones, evidencia y metadatos. Los estados iniciales de evidencia son desconocida, observada, documentada, hipótesis y validada. Validada no significa causalidad demostrada.

## Paquete relacion
La primera entrega introduce un paquete sin dependencias externas. Registra nodos y relaciones, valida extremos, consulta relaciones y calcula perfiles descriptivos. El perfil no asigna niveles de abstracción.

## Fases
1. Modelo aislado y pruebas.
2. Relaciones opcionales en OrgDecl/manifiestos y sintaxis LISPAI-AOS.
3. Grafo dentro del ensamblaje y auditoría de cambios.
4. Evaluación por Mind/Zyrion con procedencia e incertidumbre.
5. Snapshots versionados, canonicalización y recuperación compatible.

## Aceptación
- Manifiestos previos siguen parseando.
- Identidad, ciclo de vida, contratos y capacidades conservan semántica.
- Referencias inválidas se rechazan.
- Suite completa y escenarios de CI pasan con la versión Go declarada en go.mod.

## Límites de esta fase
No implementa clasificación automática de niveles, causalidad, conciencia, inferencia general ni persistencia distribuida del grafo. Esas capacidades requieren contratos y validación propios.


## Fase 2 — Relaciones declaradas y ensamblaje (implementada en la rama)

La declaración de un organismo admite ahora una lista opcional `relations`. En LISPAI-AOS se declara dentro del organismo propietario:

```lisp
(orge inventario
  (relacion inventario-almacen
    (destino almacen)
    (tipo dependencia-operativa)
    (direccion dirigida)
    (temporalidad estable)
    (evidencia documentada)
    (condiciones "solo durante operación")))
```

El origen se deriva del OID del organismo que contiene la declaración; no se duplica ni puede apuntar silenciosamente a otro organismo. La resolución registra primero todos los nodos y después valida las relaciones, por lo que una referencia puede apuntar a un organismo declarado más adelante. Un destino inexistente, un ID duplicado o una relación mal formada impiden el ensamblaje.

Las relaciones forman parte del CID canónico de la definición cuando están declaradas. La ausencia de relaciones conserva la representación canónica anterior: `nil` y una lista vacía son equivalentes y no alteran el CID. El grafo resuelto vive en `Assembly.Relations`; no sustituye contratos, bindings, estado reactivo ni capacidades.

Esta fase no interpreta automáticamente semántica causal, no concede permisos y no convierte toda relación en dependencia de ejecución. La evidencia se conserva como metadato explícito.


**Validación de esta entrega:** GitHub Actions ejecutó correctamente `go test ./...` y todos los escenarios existentes del workflow (organismo declarativo, incierto, persistencia, alset-node, autonomía, kernel y verificación de archivos de estado). La validación corresponde al commit `5aba2f1ef8b390a9ddff55192666332a7f14cdf4`; la rama continúa separada de `main`.


## Fase 3 — Contexto relacional en Runtime

El Runtime expone ahora una vista de solo lectura del grafo del ensamblaje:
- `RelationalContext(oid)` devuelve relaciones salientes y entrantes del organismo.
- `RelatedOrganisms(oid)` devuelve vecinos directos, ordenados y sin duplicados.
- `RelationalProfile()` devuelve el perfil descriptivo del grafo.
- Al cargar el ensamblaje se registra un evento `relation_graph_loaded` con el número de nodos y relaciones.

Estas consultas no escriben estado, no disparan reglas, no cambian capacidades y no se registran como mutaciones. Las relaciones entrantes se obtienen inspeccionando el snapshot del grafo, por lo que el contrato no depende de añadir semántica de recorrido al paquete `relacion`.

**Límite deliberado:** todavía no se infiere qué átomos de estado dependen de una relación. El modelo actual no define un vínculo formal entre rutas de estado y relaciones; inventarlo a partir de nombres o de la mera proximidad entre organismos sería introducir causalidad no declarada. La invalidación reactiva sigue dependiendo exclusivamente de `ReactiveState.RegisterDeps` y las reglas ARL existentes.


**Validación de Fase 3:** GitHub Actions completó con éxito `go test ./...` y el workflow de escenarios existentes en el commit `a795ee328847c53e060c389b1ac47859760d1fc6`. La suite incluye pruebas nuevas para contexto entrante/saliente, vecinos ordenados, organismo sin relaciones, organismo desconocido y ausencia de concesión de capacidades por el mero hecho de estar relacionado. El workflow terminó con conclusión `success`.


## Fase 4 — Evaluación recursiva Mind/Zyrion

Se implementa la primera integración de Mind con el grafo relacional:
- `Runtime.RulesFor` devuelve copias de las reglas de un organismo.
- `Runtime.EvaluateRuleReadOnly` inspecciona una regla sin escribir átomos, dependencias, historial ni auditoría.
- Para esta inspección contextual, primero se busca el átomo local `<oid>.<path>`; si no existe, se consulta el path global legado. Las reglas ARL normales no cambian su resolución de paths.
- `Mente.EvaluarFractal` evalúa una regla en el organismo raíz. Si Zyrion devuelve `2`, explora únicamente relaciones declaradas con el tipo explícito `contexto-fractal`; el organismo relacionado debe declarar exactamente la misma regla para participar.
- La traza conserva OID, profundidad, estado, evidencias, motivo e hijos.
- La convergencia solo acepta unanimidad de resultados determinados. Las relaciones comunes (por ejemplo, colaboración o asociación) no activan recursividad. Desacuerdo, incertidumbre, ciclos, ausencia de reglas equivalentes o límites conservan el estado `2`.
- Se aplican límites predeterminados de profundidad y cantidad de organismos para impedir expansión ilimitada.

La inspección es de solo lectura. No aplica `ThenPath`, no ejecuta genes, no cambia capacidades y no convierte conclusiones en autorizaciones. La ejecución continúa bajo los mecanismos existentes de capacidad y política.

**Alcance:** esta fase no modifica el comportamiento de `EvalRule` ni la estrategia heredada `Mente.Decidir`. La integración fractal es una API explícita que puede probarse antes de conectarla a flujos automáticos. La semántica de convergencia es una política inicial conservadora y debe permanecer visible y versionada.


## Fase 5 — Persistencia versionada y recuperación verificable

La persistencia del Runtime deja de ser un volcado JSON sin contrato explícito. El archivo `runtime_state.json` adopta una envoltura versionada con:
- Versión de esquema (`version=1`).
- Algoritmo de integridad (`sha256`) y checksum del payload canónico.
- Huella determinista del ensamblaje y mapa de RootCID por OID.
- Átomos, dependencias reactivas, invalidaciones pendientes e historial ARL.
- Auditoría, bindings y capacidades declaradas (estas últimas solo se verifican; nunca se restauran como autoridad).
- Nonces Pulse ya utilizados, para conservar la protección anti-replay después de un reinicio.
- Registros de evaluaciones fractales, con esquema, organismo raíz, regla, estado, payload y digest propio.

### Recuperación y compatibilidad

`Restore` requiere que `LoadAssembly` haya cargado previamente el ensamblaje. Antes de modificar el Runtime valida:
1. La versión y el algoritmo de checksum.
2. La integridad del payload.
3. La coincidencia exacta de los RootCID de definición y de la huella del ensamblaje.
4. La igualdad de bindings y capacidades frente al ensamblaje activo.
5. La integridad y pertenencia de los registros de evaluación.

La recuperación reemplaza átomos, dependencias, invalidaciones e historial con el snapshot validado y restaura los nonces ya consumidos para evitar replay; no mezcla el snapshot con datos residuales de una ejecución anterior. La auditoría recuperada recibe un nuevo evento `restore`.

Las capacidades y RootCID **no se restauran desde el archivo**: la autoridad procede de las declaraciones que se cargaron en el Runtime. El snapshot solo permite comprobar que no han cambiado. Por ello un archivo manipulado no puede otorgar permisos.

Se admite migración controlada del formato JSON anterior sin envoltura, pero únicamente si incluye RootCID que coinciden exactamente con el ensamblaje cargado. Los snapshots antiguos sin RootCID se rechazan: no existe evidencia suficiente para asociar sus datos a una definición concreta. Tras recuperar un formato legado, el siguiente `Persist` lo escribe en el formato versionado actual.

### Evaluaciones fractales

`Mente.EvaluarFractal` registra una traza de evidencia `mind.fractal.v1` en el Runtime. Este registro es metadato de evaluación: no modifica átomos de dominio, historial ARL, auditoría, capacidades ni ejecuta genes. El registro conserva el payload completo y un digest SHA-256, y viaja dentro del snapshot para permitir inspección posterior.

### Límites explícitos

- El checksum detecta corrupción accidental o alteración no coordinada del payload; **no es una firma de autoría ni protege frente a un atacante capaz de reemplazar simultáneamente el contenido y su checksum**.
- Esta fase no implementa replicación distribuida, consenso, disponibilidad de bloques IPFS ni recuperación entre peers.
- RootCID identifica la definición canónica del organismo; no identifica una instancia viva, un nodo/peer ni garantiza que el contenido esté disponible.
- La identidad criptográfica del nodo y la disponibilidad del contenido continúan siendo contratos separados.

### Pruebas añadidas

Se cubren round-trip, versión y checksum, corrupción sin restauración parcial, incompatibilidad de RootCID, migración del formato legado, rechazo de capacidades manipuladas y conservación de trazas fractales. La validación integral queda condicionada al resultado de GitHub Actions de este commit.


## Fase 6 — Bloques verificables y persistencia P2P local

Esta fase conecta el RootCID con almacenamiento direccionado por contenido y con el transporte P2P existente, sin cambiar el contrato semántico de los organismos.

### Bloques raw direccionados por contenido

Se añade `p2p.AlmacenContenido`:
- Acepta un bloque solo si `rootcid.Verificar(bytes, rootCID)` confirma su digest.
- Usa el digest hexadecimal validado como nombre de almacenamiento; nunca una ruta suministrada libremente.
- Los bloques son inmutables: repetir el mismo contenido es idempotente y una colisión/sobrescritura se rechaza.
- Publica escrituras mediante archivo temporal, sincronización y rename atómico.
- Verifica de nuevo el contenido en cada lectura.
- Devuelve `ErrContenidoNoDisponible` cuando el bloque no existe; un archivo presente pero corrupto no cuenta como disponible.

`CIDRoot` representa el digest RootCID como CIDv1, codec `raw`, multihash SHA2-256. Por ello los bloques son compatibles con la identificación de contenido raw de IPFS. El almacén implementado es local y basado en archivos; no debe confundirse con un nodo IPFS completo ni con Bitswap.

### Integración con el nodo P2P

- `NuevoConEstado` inicializa el almacén de bloques y verifica las entradas históricas del índice `organismos.json`. Un índice que referencia contenido corrupto impide el arranque.
- `AnunciarOrganismo` valida y persiste primero el bloque, después actualiza el índice durable y finalmente anuncia el proveedor en Kademlia.
- `RecuperarOrganismo` consulta primero el bloque local; si no está disponible, busca proveedores, verifica la respuesta y la cachea localmente.
- `recibirOrganismo` solo sirve bytes que el almacén vuelve a verificar.
- `ReanunciarOrganismos` comprueba la disponibilidad e integridad local antes de publicar cada proveedor.
- `DisponibleLocal` expresa exclusivamente disponibilidad local verificada; no es una afirmación sobre la red.

### Persistencia Kademlia

El ejecutable `cmd/alset-p2p` activa Pebble por defecto en `estado/dht`, configurable con `-dht`. El datastore temporal se cierra al cambiar a Pebble y el cierre del nodo intenta cerrar DHT, datastore y Host aun cuando una de esas operaciones falle.

Los registros de proveedor Kademlia no son una garantía perpetua: expiran y requieren renovación. El nodo reanuncia organismos verificados después de conectarse al peer indicado para bootstrap. La persistencia del datastore no equivale a disponibilidad global.

### Contratos de identidad y estado

- RootCID sigue identificando contenido de definición; no identifica instancia ni peer.
- NodeID/PeerID siguen identificando el nodo de transporte mediante su identidad persistente.
- La disponibilidad es un hecho verificable y separado de ambos.
- Se conserva el contrato Zyrion: `1=True`, `0=False`, `2=Indeterminado`. El estado `2` activa el siguiente nivel fractal con las mismas reglas, sujeto a límites explícitos; no se altera la semántica ternaria en esta fase.

### Límites que permanecen abiertos

No se implementan en esta fase Bitswap, GossipSub, consenso, replicación automática basada en políticas, reparación de bloques, garbage collection, retención por pinning ni recuperación multi-peer con reconciliación de versiones. Tampoco se afirma que la DHT por sí sola garantice disponibilidad. La próxima fase debe definir el protocolo de replicación y sus garantías antes de automatizarlo.

### Pruebas incorporadas y validación

Se prueban almacenamiento persistente, idempotencia, rechazo de RootCID incorrecto, detección de corrupción al leer y al arrancar, reinicio del nodo, cacheo de contenido recuperado por DHT y reapertura del datastore Pebble. GitHub Actions completó con éxito `go test ./...` y todos los escenarios del workflow (organismo declarativo, incierto, persistente, alset-node, autonomía, arranque del kernel y verificación de estado) para el commit `4e4500aa9ad54042f9f7c319dc0e54e9cd7b7ea9`. Run: https://github.com/yecharlot/AlsetOS/actions/runs/43690329073.
