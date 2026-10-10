# AlsetOS Genesis — guía de trabajo

## Qué es

Kernel propio orientado a **ORGES** y al modelo del manifiesto ATS:

- Unidad fundamental: **organismo** (OCB), no proceso Unix.
- Comunicación: **Pulse** (con capacidades).
- Seguridad: **default-deny** + capacidades explícitas.
- Lógica: **Zyrion** (V / F / I; la I no se fuerza a falso).
- Nodo: **soberano local** (NodeID), sin registro central.
- Interfaz: **shell declarativa** (comandos legibles).

No imita Linux ni Windows. No hay usuarios POSIX ni “todo es un fichero”.

## Arranque

```bash
cd kernel_native
make clean && make && make run
```

Clic en la ventana QEMU para el teclado. Escribe `help` y Enter.

## Organismos al nacer

| Nombre    | Rol |
|-----------|-----|
| Master    | Soberano del **nodo** (todas las caps). No es “root Unix”. |
| Shell     | Actor por defecto de la consola (least privilege). |
| PulseBus  | Infraestructura de mensajería. |

## Comandos

```
help
node
organisms          (o ls)
actor Shell
actor Master
create orges fin seguridad alerta
create peer nodoB
pulse PulseBus
pulse fin ping
grant fin pulse.send
caps
zyrion V and I
zyrion not I
goals fin
clear
```

### Seguridad en la práctica

1. Con actor `Shell` puedes crear ORGES (tiene `orges.create`).
2. Cambia a un ORGES sin `pulse.send` y `pulse` fallará → **default-deny**.
3. `actor Master` luego `grant <orges> pulse.send` → otorgas capacidad.
4. Vuelve al ORGES y el Pulse ya puede enviarse.

Eso es el modelo de capacidades del manifiesto, en miniatura y usable.

### Zyrion

```
zyrion V and I     → I
zyrion V or I      → V
zyrion not I       → I
```

La incertidumbre permanece incertidumbre.

## Cómo pensar un ORGES

1. **Nombre** legible (`fin`, `triaje`, `logistica`).
2. **Goals** en el create: `create orges fin seguridad cumplimiento`.
3. **Caps mínimas**: al crear recibe send/recv; el resto se **grant**.
4. **Pulse** solo entre organismos que pueden enviar/recibir.
5. Un **Peer** es otro organismo de malla, no una “cuenta de usuario”.

## Relación con el Desktop rico (Go)

- **Genesis kernel**: dominio del modelo y de la seguridad en metal (QEMU).
- **alsetos-native**: UI rica y export PWA para diseñar ORGES visualmente.

Misma filosofía; dos superficies. El objetivo de producto es que ORGES nacidas aquí o en Studio compartan identidad, caps y Pulse.

## Límites actuales (honestos)

- Shell texto, no ventanas compositadas.
- Pulse local al nodo (aún no red entre dos QEMU).
- Caps en bits en RAM (sin persistencia tras reinicio).
- Firmas criptográficas fuertes: en el runtime Go de referencia; aquí la autoridad es el chequeo de caps + nonce.

Aun así ya puedes **dominar** el ciclo: crear ORGES, denegar, otorgar, evaluar Zyrion, listar, cambiar actor.

## Frase de diseño

Si algo no es un organismo con capacidades y Pulse, no pertenece al núcleo de AlsetOS Genesis.
