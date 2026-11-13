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
