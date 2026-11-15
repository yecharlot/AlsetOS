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


## Fase 2 — Relaciones declaradas y ensamblaje (en implementación)

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
