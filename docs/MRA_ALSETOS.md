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


## Fase 7 — Replicación versionada y reconciliación de linajes

La fase 7 añade una primera replicación explícita entre peers conectados y un contrato verificable para describir versiones de un mismo organismo lógico. No se interpreta el RootCID como identificador lógico estable: cada versión tiene su propio RootCID y el campo `OrganismoID` identifica el linaje.

### Descriptor firmado de versión

`p2p.DescriptorVersion` contiene:
- Schema del descriptor.
- `OrganismoID` estable definido por el dominio.
- Número de revisión.
- RootCID del contenido inmutable.
- RootCID de la revisión antecesora.
- NodeID y PeerID del autor.
- Clave pública, fecha y firma Ed25519.

La firma cubre todos los campos del descriptor. La verificación comprueba que NodeID y PeerID derivan de la misma clave pública y que la firma es válida. Esto acredita integridad y autoría criptográfica; **no** acredita que el autor esté autorizado, sea confiable o pueda ejecutar el organismo.

### Replicación explícita

El protocolo `/alset/version/1.0.0` transporta el descriptor y los bytes del bloque:
1. El emisor verifica descriptor y RootCID.
2. El receptor vuelve a verificar firma, identidad declarada y RootCID.
3. El bloque se almacena de forma inmutable.
4. El descriptor se registra en el índice versionado.
5. El receptor confirma RootCID y estado del linaje.

Se añade `ReplicarVersionAConocidos` para realizar fan-out explícito a los peers que ya están conectados. Los resultados se devuelven individualmente por peer. No se descubre una topología nueva ni se declara retención futura a partir de una confirmación.

### Reconciliación sin sobrescritura

El índice `versiones.json` conserva descriptores firmados y se escribe de forma atómica. Las clasificaciones posibles son:
- `coherente`: antecesores disponibles y sin bifurcaciones observadas.
- `incompleto`: falta el antecesor de una revisión.
- `conflicto`: dos RootCID distintos para una misma revisión, ramas divergentes desde un mismo antecesor o declaraciones incompatibles del antecesor.

Los conflictos no se resuelven por fecha, orden de llegada ni last-write-wins. Se conservan ambas evidencias. La llegada posterior de una revisión no borra un conflicto anterior. El sistema no elige un ganador ni concede autorización.

La recepción de una revisión fuera de orden puede almacenar su bloque y descriptor, pero el linaje queda incompleto hasta recuperar su antecesor. La recepción de una bifurcación conserva las dos ramas y marca conflicto.

### Límites de esta fase

Esta implementación utiliza un protocolo libp2p de petición/respuesta y fan-out explícito a peers conectados. **No es GossipSub** ni un mecanismo epidémico automático de eventos. Quedan pendientes:
- GossipSub y sus políticas de validación de mensajes.
- Descubrimiento y membresía de topics.
- Replicación automática por factor de redundancia y política de retención.
- Recuperación de bloques desde múltiples proveedores en paralelo.
- Reparación, pinning, garbage collection y pruebas de retención.
- Política de confianza/autorización para autores y validadores.
- Resolución humana o mediante reglas explícitas de conflictos; nunca LWW implícito.

No se introduce consenso distribuido ni CRDT: el contrato de versiones no los necesita para detectar y preservar bifurcaciones.

### Pruebas incorporadas

Se prueban firma y derivación de NodeID/PeerID, alteración de descriptor, idempotencia, revisión sin antecesor, llegada posterior del antecesor, fork de una revisión, persistencia del conflicto tras reinicio y replicación entre dos nodos con conservación de ambas versiones.


### Validación CI de la Fase 7

GitHub Actions completó con éxito para el commit `2ea50ea49e45aa7121c254ed96a3ab861ff3776e`: `go test ./...` y todos los escenarios de integración del workflow, incluidos los escenarios de organismo, autonomía y kernel. Ejecución: https://github.com/yecharlot/AlsetOS/actions/runs/43693664517.


## Fase 8 — Anuncios de versiones mediante GossipSub

Esta fase incorpora un topic GossipSub para propagar anuncios de versiones firmadas. El protocolo de transferencia de bloques de la fase 7 se mantiene separado.

### Contrato del evento

- Topic: `alset/versions/1.0.0`.
- Schema del evento: `alset.version-announcement.v1`.
- El cuerpo contiene el `DescriptorVersion` firmado; no contiene los bytes del organismo.
- El tamaño máximo del evento de aplicación es 64 KiB.
- GossipSub exige firma de mensaje y verificación estricta de la identidad libp2p que publica el evento.
- El topic activa `WithFloodPublish` para que los anuncios pequeños lleguen a todos los peers suscritos conocidos, incluso en redes iniciales pequeñas. Es una decisión de difusión para esta etapa; debe revisarse con métricas al crecer la red porque el coste aumenta con el número de peers.
- Un validador síncrono comprueba estructura, schema, tamaño y firma Ed25519 del descriptor antes de aceptar y reenviar el anuncio.
- Se rechazan mensajes malformados, campos desconocidos, datos JSON adicionales, schema no soportado y firmas inválidas.

La firma del mensaje GossipSub acredita qué identidad libp2p publicó el anuncio. La firma interna del descriptor acredita la autoría de la versión. Son comprobaciones distintas. La retransmisión de un anuncio no transforma al relay en autor ni en autoridad de confianza.

### Recepción y reconciliación

Los nodos suscritos consumen los eventos aceptados y registran el descriptor en `versiones.json`. El registro es idempotente. La recepción fuera de orden conserva el estado `incompleto`; la recepción de ramas incompatibles conserva `conflicto`. Ningún evento selecciona una rama ganadora ni concede capacidades.

El anuncio es una pista de descubrimiento y sincronización, no una prueba de disponibilidad. El evento no transporta el bloque ni afirma que el emisor lo conserve. Para obtener el contenido, el consumidor debe recuperarlo por su RootCID mediante los mecanismos de almacenamiento y recuperación existentes y verificarlo antes de usarlo. La propagación GossipSub tampoco garantiza que un nodo desconectado reciba eventos publicados durante su ausencia.

### Ciclo de vida y límites

Cada nodo inicia el servicio GossipSub, valida y se suscribe al topic de versiones. Al cerrar el nodo se cancela el contexto de GossipSub y se detiene el consumidor de eventos. Los errores de persistencia del índice se reportan en el log local; no se presenta un evento como reconciliado si no pudo registrarse.

Esta fase **no** implementa todavía:
- Recuperación automática de bloques al recibir un anuncio.
- Sincronización histórica completa de descriptores para peers que estuvieron desconectados.
- Factor de redundancia, selección de réplicas o política de retención/pinning.
- Reparación automática, garbage collection ni disponibilidad garantizada.
- Autorización o confianza basada únicamente en la firma.
- Resolución automática de forks, consenso o CRDT.

### Pruebas previstas para esta entrega

La suite añade escenarios de propagación entre tres nodos, rechazo de anuncios malformados y firmas alteradas, recepción fuera de orden, deduplicación idempotente y conservación de bifurcaciones. Las pruebas de malla usan hosts libp2p aislados para no depender del descubrimiento mDNS ni del DHT en esta validación. Se comprueba además que recibir un anuncio no hace aparecer el bloque en el almacén local.\n\nGitHub Actions completó con éxito el commit de implementación `7a4daae14b1190d6215789df884c897dbf72f8cf`, incluyendo `go test ./...`, escenarios de integración, autonomía, arranque del kernel y verificación de estado: https://github.com/yecharlot/AlsetOS/actions/runs/43695107823. La rama sigue sin fusionarse.


## Fase 9 — Recuperación verificable desencadenada por anuncios

La recepción de un anuncio GossipSub válido ya no se limita a registrar el descriptor. El nodo incorpora una cola acotada de recuperación asíncrona:
1. Registra el descriptor firmado en el índice durable.
2. Comprueba si el bloque RootCID ya está disponible localmente.
3. Si falta, intenta obtenerlo del peer que entregó el anuncio. Ese peer es solo un primer candidato y puede ser un relay sin el contenido.
4. Si no lo tiene, consulta proveedores Kademlia mediante los mecanismos existentes.
5. Verifica nuevamente RootCID antes de almacenar el bloque.
6. Si la DHT está disponible, anuncia el nodo como proveedor después de guardar el contenido verificado.

La cola tiene capacidad limitada y cada recuperación dispone de un timeout. Si se llena, el anuncio permanece registrado pero la recuperación se descarta y se informa; puede reintentarse mediante una nueva sincronización o una operación explícita. El bloque no se ejecuta ni concede capacidades por el hecho de haberse recuperado.

La publicación de una versión local registra el proveedor Kademlia después de guardar el bloque y antes de publicar el anuncio GossipSub. La replicación entrante también intenta publicar el bloque verificado como proveedor. La firma del descriptor sigue siendo procedencia, no autorización.

**Límites:** no se garantiza que exista un proveedor disponible; no hay sincronización histórica para eventos perdidos, factor de redundancia, retención/pinning, reparación, selección de confianza ni recuperación paralela multi-peer. La recuperación automática está limitada por tamaño, cola y tiempo, pero aún requiere una política de confianza y cuotas más completa para redes abiertas.

**Pruebas:** la prueba de malla de tres nodos verifica que un anuncio firmado provoca recuperación real del bloque y que el bloque almacenado supera el RootCID esperado. La suite conserva pruebas de rechazo de anuncios malformados, firmas alteradas, deduplicación, orden fuera de secuencia y forks.


## Fase 10 — Sincronización de versiones al reconectar y recuperación por RootCID

La Fase 10 cierra la brecha entre los anuncios efímeros GossipSub y los nodos que estuvieron desconectados. Al establecerse una conexión libp2p, cada nodo solicita al peer remoto una instantánea de su inventario de descriptores mediante `/alset/version-sync/1.0.0`.

### Intercambio de inventarios

- El inventario contiene descriptores de versión firmados, no los bytes de los organismos.
- La respuesta está limitada a 10 000 descriptores y 8 MiB.
- Cada descriptor se verifica criptográficamente antes de registrarse.
- El registro conserva idempotencia, linajes incompletos y bifurcaciones.
- Una sincronización puede repetirse; no altera el significado ni la autoridad de una versión ya registrada.
- El protocolo de inventario es independiente de GossipSub: los anuncios aceleran el descubrimiento en línea; el inventario al reconectar recupera el historial que el nodo pudo perder.

### Recuperación de contenido

Después de registrar cada descriptor válido, el nodo encola su RootCID en el trabajador de recuperación existente. Este consulta primero al peer del inventario y verifica los bytes contra RootCID antes de almacenarlos. Si el peer no conserva el bloque, el mecanismo intenta los proveedores Kademlia conocidos. Un bloque recuperado se puede anunciar como proveedor, pero la publicación DHT es de mejor esfuerzo y no invalida el contenido local si la tabla de encaminamiento aún no está lista.

### Límites y semántica

- La sincronización no implementa consenso, CRDT ni resolución automática de forks.
- No hay política de confianza o autorización: firma válida significa procedencia verificable, no permiso.
- Un descriptor no prueba disponibilidad; un inventario completo no garantiza que todos sus bloques sigan existiendo.
- La cola de recuperación tiene capacidad 64; si se llena, se registra el descarte de esa solicitud. La sincronización puede repetirse para reencolar bloques pendientes.
- La respuesta de inventario es una instantánea completa acotada, no paginada. El límite impide crecimiento ilimitado; si se supera, el peer debe usar una futura estrategia de paginación.
- No se introduce retención, pinning, factor de redundancia, reparación ni descarga paralela multi-peer.

### Pruebas

Se incorpora una prueba de reconexión: un nodo publica una versión mientras el otro está desconectado; al reconectar, el segundo obtiene el descriptor desde el inventario y recupera los bytes por RootCID, comprobando integridad y estado de linaje. También se prueba la sincronización idempotente y la conservación de forks. La suite debe validar que no se ejecuta el organismo por el mero hecho de sincronizar o recuperar su contenido.


## Fase 11 — Prueba de disponibilidad y redundancia verificable

Esta fase añade una primera medición operativa de disponibilidad que no confunde anuncios DHT, inventarios o descriptores firmados con posesión real del contenido.

### Recibo de disponibilidad

El protocolo `/alset/content-availability/1.0.0` usa un desafío aleatorio del solicitante. El peer remoto responde únicamente si `AlmacenContenido.Disponible` puede leer y verificar el RootCID. El recibo incluye schema, RootCID, PeerID, clave pública, nonce, emisión, expiración y firma Ed25519. El solicitante verifica que:
- el recibo corresponde al RootCID y al desafío recién enviado;
- el PeerID deriva de la clave pública incluida;
- la firma cubre todos los campos;
- la ventana temporal no está vencida ni excede el máximo permitido.

El desafío evita reutilizar una respuesta antigua como prueba de una consulta nueva. El recibo es evidencia temporal de disponibilidad local, no promesa de retención, reputación, confianza o autorización.

### Factor de redundancia

`AsegurarRedundancia(ctx, descriptor, contenido, factor)` recibe una versión explícita, verifica su firma y RootCID, exige que el coordinador conserve su propia copia y replica a peers conectados de forma determinista. El factor permitido es 1..32 e incluye la copia local. Cada copia remota solo se contabiliza después de una consulta de disponibilidad con recibo válido y fresco. La operación devuelve el número de copias verificadas y resultados individuales; si no alcanza el objetivo, informa redundancia insuficiente sin ocultar el progreso parcial.

La transferencia puede registrar el descriptor aunque el linaje remoto quede incompleto o en conflicto; eso no invalida por sí solo la copia del bloque. Por ello la disponibilidad se comprueba mediante el protocolo independiente y no se deduce del estado del linaje.

### Límites deliberados

- Solo se consideran peers conectados en el momento de la operación.
- La replicación es secuencial; no se añade todavía descarga/subida paralela.
- No hay retención/pinning persistente, reserva de almacenamiento, renovación de recibos ni reparación automática tras una caída.
- Un recibo firmado no impide que el peer borre el bloque inmediatamente después.
- No hay política de confianza o autorización; firma válida acredita procedencia criptográfica.
- No se introduce consenso, CRDT ni resolución automática de forks.
- La disponibilidad de una versión concreta no implica que esa versión sea la autorizada para ejecutar.

### Pruebas

Se prueba que un peer sin bloque no puede emitir un recibo válido, que la replicación eleva el factor solo tras confirmación firmada, y que se rechazan recibos caducados o ligados a otro desafío. La CI debe ejecutar estas pruebas junto con la suite completa y los escenarios de integración existentes.


## Fase 12 — Retención persistente, pinning y reparación

### Política de retención

Se incorpora `GestorRetencion`, con persistencia atómica en `retenciones.json`. Cada registro conserva schema, descriptor firmado, RootCID, factor solicitado y marcas temporales. La política es local al nodo; no se interpreta como una orden remota ni como una concesión de confianza.

`Nodo.FijarVersion(descriptor, factor)` exige:
- Descriptor Ed25519 válido.
- Descriptor previamente registrado en el índice local de versiones.
- Bloque presente y verificable localmente.
- Factor entre 1 y 32, incluyendo la copia local.

`LiberarVersion(rootCID)` elimina la política, no el contenido. El almacén no tiene recolección de basura en esta fase; por tanto, liberar un pin no elimina datos ni invalida descriptores.

### Reparación verificable

`RepararRetenciones(ctx)` recorre las políticas persistidas. Si el bloque local falta o está corrupto, solicita bytes a peers conectados y después a proveedores Kademlia. Cada respuesta se valida contra el RootCID antes de usar `AlmacenContenido.Reparar`, que reemplaza solo contenido ausente o corrupto mediante temporal, fsync y rename; un bloque ya íntegro nunca se sobrescribe.

Una vez restaurada la copia local, se llama a `AsegurarRedundancia` para recuperar el factor solicitado. Los peers se verifican con recibos firmados y frescos. Las bifurcaciones no se resuelven: la reparación usa el descriptor explícito fijado por la política.

### Mantenimiento periódico

`IniciarMantenimiento(intervalo)` permite activar un worker periódico de manera explícita. El intervalo mínimo es un minuto y cada ciclo tiene timeout de 30 segundos. El worker no se inicia automáticamente: cada despliegue decide cuándo asumir consumo de red, CPU y disco. `Cerrar` cancela y espera el worker antes de cerrar los servicios P2P.

### Límites y recuperación

- La retención es una política local; no reserva espacio ni obliga a peers remotos a conservar bloques.
- La liberación del pin no borra datos.
- La reparación necesita al menos una fuente íntegra accesible; si no existe, informa fallo y conserva la política para futuros intentos.
- No hay cuotas, reputación, penalización, consenso ni resolución de forks.
- No se promete disponibilidad futura por haber recibido un recibo.
- La reparación de una copia corrupta solo ocurre después de verificar el RootCID del contenido sustituto.

### Pruebas

Se añaden pruebas de persistencia/liberación de políticas, reparación de bloque corrupto, recuperación de réplica remota desaparecida, recuperación de copia local corrupta desde otro peer y control del intervalo/duplicación del worker periódico.


## Fase 13 — Compromisos distribuidos de retención

### Contrato del lease

Se incorpora el protocolo `/alset/retention-lease/1.0.0`. Un solicitante que ya replicó el contenido puede pedir a un proveedor un compromiso temporal. La petición se firma con la identidad del solicitante y el proveedor comprueba que la clave pública corresponde al PeerID de la conexión. El proveedor vuelve a leer el bloque y verifica su RootCID antes de aceptar. La respuesta contiene un lease firmado por el proveedor, vinculado a RootCID, OrganismoID, solicitante, proveedor, tamaño, emisión y expiración.

La firma del descriptor demuestra procedencia; no autoriza almacenamiento remoto. La conexión y la identidad del peer tampoco autorizan por sí solas la solicitud.

### Política de aceptación

El proveedor opera en denegación por defecto. La aceptación exige configuración local explícita mediante `AutorizarRetencionRemota(peerID, cuotaBytes)` y `ConfigurarCapacidadRetencion(bytes)`. Se contabilizan los leases vigentes para aplicar la cuota individual y el límite global. Si falta configuración, la solicitud se rechaza.

Las autorizaciones y la capacidad son configuración operativa en memoria: después de un reinicio el proveedor vuelve a denegar solicitudes hasta que el operador las configure. Los leases aceptados se guardan de forma persistente y atómica en `leases-retencion.json`.

### Renovación, liberación y expiración

- `SolicitarRetencionRemota`: solicita un lease temporal después de que el bloque ya se haya replicado.
- `RenovarLeaseRemoto`: renueva un compromiso conocido sin permitir alterar solicitante, proveedor, OrganismoID o RootCID.
- `LiberarLeaseRemoto`: libera el compromiso de forma explícita.
- `ExpirarLeases`: elimina compromisos vencidos del registro.
- El worker opt-in de mantenimiento ejecuta la expiración periódica junto con la reparación de contenido.

La liberación y la expiración no borran los bytes. No se introduce garbage collection ni se interpreta la liberación como autorización para borrar un bloque que pudiera estar referenciado por otras versiones.

### Redundancia contractual

Se modifica `AsegurarRedundancia`: cuenta la copia local y solo leases remotos aceptados, firmados y vigentes. La replicación exitosa o un recibo de disponibilidad puntual no son suficientes. La operación automática solicita leases de 24 horas; el protocolo admite duraciones explícitas de hasta 30 días.

### Límites y garantías

- Un lease firmado acredita que el proveedor asumió un compromiso temporal verificable; no garantiza que el hardware no falle, que el proceso siga activo o que el proveedor no incumpla.
- La capacidad es declarada por el operador; todavía no se mide automáticamente el espacio libre del sistema de archivos.
- No hay reputación, pago, penalización, consenso ni resolución de disputas.
- Las autorizaciones/cuotas no se infieren de la firma del descriptor ni se restauran automáticamente tras reinicio.
- No se borra contenido al liberar o expirar un lease.
- La firma acredita compromiso y procedencia, no confianza absoluta ni autorización para ejecutar organismos.

### Pruebas

Se incorporan pruebas de denegación por defecto, aceptación tras autorización explícita, validación de firma, persistencia, cuota/capacidad, renovación, liberación, expiración y contabilización de redundancia mediante leases vigentes. La suite de CI debe verificar también los escenarios previos de persistencia, reparación y replicación P2P.
