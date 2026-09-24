# Contrato del Núcleo de AlsetOS v0.1

## Propósito

AlsetOS es un sistema operativo adaptativo orientado a organismos digitales persistentes.

Su modelo fundamental es:

`programa → proceso → ejecución`

frente a:

`descripción → RootCID → organismo → Genes + Agentes + Memoria → LispAI + Mind + Zyrion → adaptación → ejecución`

Tiny Core/Linux actúa como sustrato mínimo de hardware. AlsetOS aporta identidad, ejecución, adaptación, coordinación y semántica del organismo.

## Componentes fundamentales

- **RootCID:** identidad persistente y direccionamiento verificable del organismo.
- **Gene:** capacidad o módulo desplegable.
- **Agente:** actor autónomo limitado por capacidades.
- **Memoria:** estado, historial, conocimiento y datos de recuperación.
- **Pulso:** mecanismo de eventos, comunicación y actualización.
- **LispAI:** intérprete y lenguaje dinámico fundamental del sistema.
- **Mind:** capa de observación, razonamiento, decisión, acción y recuperación.
- **Zyrion:** lógica nativa ternaria 0/1/2 para estados, incertidumbre y evaluación.
- **Capacidad:** permiso explícito para utilizar recursos.
- **Nodo Alset:** lugar físico o virtual donde vive un organismo.

## Arquitectura

```
APLICACIONES / ORGANISMOS
        │
ORGANISMO RUNTIME
RootCID · Genes · Agentes · Memoria · Pulso · Capacidades
        │
MIND
Observación · Razonamiento · Decisión · Acción · Recuperación
        │
LISPAI
Interpretación · Metaprogramación · Composición · Ejecución
        │
ZYRION
Lógica 0/1/2 · Incertidumbre · Evaluación
        │
NODO ALSET
Identidad · Red · Almacenamiento · Dispositivos · Seguridad
        │
TINY CORE / LINUX
        │
HARDWARE
```

## Reglas

1. La identidad pertenece al organismo, no al nodo.
2. Un organismo debe poder recuperarse en otro nodo compatible.
3. Ningún Agente puede utilizar una capacidad que no le haya sido concedida.
4. LispAI puede interpretar y modificar comportamiento dentro de los límites de seguridad y capacidades.
5. Mind decide; las capacidades autorizan.
6. Zyrion complementa la lógica convencional; no la reemplaza obligatoriamente.
7. Los cambios relevantes deben poder quedar registrados en Memoria y Pulso.
8. Los Genes deben declarar identidad, versión, capacidades e interfaz.
9. La implementación inicial debe favorecer portabilidad y aislamiento, especialmente mediante WASM cuando sea apropiado.
10. No añadir una nueva abstracción si puede expresarse naturalmente mediante RootCID, Gene, Agente, Memoria, Pulso, Capacidad, LispAI, Mind o Zyrion.

## Aplicaciones como organismos

Una aplicación AlsetOS no se considera únicamente un proceso. Es un organismo compuesto por identidad, capacidades, código, agentes, memoria, comunicación y estado persistente.

Ciclo mínimo:

`crear → identificar → verificar → desplegar → ejecutar → recordar → adaptar → recuperar`

## Multilingüismo

El repositorio inicial utiliza **español** para variables, estructuras, nombres conceptuales y comentarios.

La arquitectura separa:

`lenguaje interno ≠ lenguaje de interfaz ≠ lenguaje de documentación`

Las futuras traducciones no deben alterar la semántica interna del núcleo.

## Objetivo

AlsetOS busca transformar el sistema operativo de un ejecutor pasivo de programas en un entorno capaz de alojar organismos digitales persistentes, adaptarse al entorno, razonar bajo incertidumbre y sobrevivir a cambios de nodo dentro de límites autorizados.
