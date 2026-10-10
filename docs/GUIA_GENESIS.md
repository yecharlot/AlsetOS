# AlsetOS Genesis — Guía completa (español)

## Visión

AlsetOS Genesis es un kernel orientado a **organismos**, no a procesos Unix.  
El hardware se modela como **organismos-dispositivo**.  
La interacción entre todo es **Pulso** sujeto a **capacidades** (default-deny).  
La lógica ternaria es **Zyrion** (V / F / I).

## Arranque

```bash
cd kernel_native
make clean && make && make run
```

Haz clic en la ventana de QEMU para que el teclado/ratón vayan a la VM.

## Organismos al iniciar

| Nombre     | Tipo        | Rol |
|------------|-------------|-----|
| Maestro    | MAESTRO     | Soberano del **nodo** (todas las capacidades) |
| Consola    | CONSOLA     | Actor por defecto de la shell |
| BusPulso   | BUS         | Infraestructura de mensajería |
| Teclado    | DISP        | Organismo-dispositivo entrada |
| Raton      | DISP        | Organismo-dispositivo puntero |
| Memoria    | DISP        | Organismo-dispositivo memoria |
| Disco      | DISP        | Ramdisk (sectores en RAM) |
| Red        | DISP        | Red local (sin enlace externo aún) |
| Extraible  | DISP        | Unidad extraíble (modelo) |

## Interfaz

### Consola (por defecto)

Línea de comandos en español. Escribe `ayuda`.

### Gráfica

```
interfaz grafica
```

o tecla **F2**. Volver a consola: **F1** o `interfaz consola`.

En gráfica:

- Panel izquierdo: organismos-dispositivo  
- Panel derecho: Maestro, Consola, ORGES, pares  
- **W/S** o flechas: cambiar selección  
- **Enter**: Pulso desde Consola al seleccionado  
- **1**: crear ORGES de demostración  
- Ratón: mover cursor `X`; clic izquierdo selecciona  

## Comandos (español)

| Comando | Abreviatura | Descripción |
|---------|-------------|-------------|
| ayuda | aid, ? | Lista de comandos |
| nodo | | NodeID y actor actual |
| organismos | lista, ls | Todos los OCB |
| dispositivos | disp | Solo dispositivos |
| actor \<nombre\> | | Cambiar organismo activo |
| capacidades | caps | Caps del actor |
| crear orges \<nom\> [metas…] | | Nuevo ORGES |
| crear par \<nombre\> | | Peer descentralizado |
| pulso \<destino\> | | Enviar Pulso |
| otorgar \<org\> \<cap\> | | Conceder capacidad |
| metas \<nombre\> | | Metas de un ORGES |
| estado \<disp\> | | Estado de dispositivo |
| leer \<disp\> [sector] | | Leer vía organismo-disp |
| escribir disco \<sec\> \<texto\> | | Escribir ramdisk |
| zyrion A y\|o\|no B | | Lógica ternaria |
| interfaz consola\|grafica | gui, cli | Cambiar UI |
| limpiar | cls | Limpiar pantalla |

### Capacidades (nombres cortos)

| Nombre completo | Abreviatura |
|-----------------|-------------|
| pulso.env | env |
| pulso.rec | rec |
| orges.crear | crear |
| orges.destr | destr |
| caps.otorg | otorg |
| nodo.admin | admin |
| zyrion | zyrion |
| disp.leer | leer |
| disp.escr | escr |
| disp.ctrl | ctrl |
| todo | todo |

## Ejemplos

```
ayuda
dispositivos
estado Disco
escribir disco 0 hola
leer Disco 0
crear orges fin seguridad
pulso BusPulso
actor Maestro
otorgar fin disp.leer
actor fin
leer Memoria
zyrion V y I
interfaz grafica
```

## Seguridad

- **Default-deny**: sin capacidad, la acción falla con mensaje claro.  
- **Maestro** del nodo puede todo; la Consola arranca con un conjunto limitado.  
- Acceso a dispositivos vía caps `disp.leer` / `disp.escr` / `disp.ctrl`.  
- Todo acceso relevante deja rastro de **Pulso** (contadores pin/pout).

## Filosofía de dispositivos

En Linux “todo es un archivo”.  
En AlsetOS **todo dispositivo es un organismo** con identidad, capacidades y Pulse.

Así el mismo modelo sirve para ORGES de negocio y para Teclado, Disco o Red.

## Límites actuales

- Disco = ramdisk en memoria (se pierde al apagar QEMU).  
- Red = organismo local, aún sin paquetes entre dos nodos.  
- Extraíble = modelo de organismo (sin USB host real).  
- UI gráfica = panel VGA texto + ratón (no compositor TrueColor).  
- Un solo idioma de comandos: español (i18n después).

## Archivos

- `kernel_native/src/kernel.c` — núcleo Genesis  
- `kernel_native/Makefile` — build y `make run`  
- `docs/GUIA_GENESIS.md` — esta guía  
