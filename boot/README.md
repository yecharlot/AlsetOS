# Arranque AlsetOS sobre TinyCore/Linux

BIOS/UEFI -> TinyCore/Linux -> /opt/alset/alset-kernel -> NodeID + DHT + Autonomía -> Organismos

El kernel AlsetOS no reemplaza el kernel Linux del medio de arranque inicial.

## Instalación mínima

Copiar el binario alset-kernel a /opt/alset/alset-kernel y el script alset-init a un lugar ejecutable por TinyCore.

El script crea el estado persistente y arranca el kernel.

## Objetivo

La primera imagen de AlsetOS puede utilizar TinyCore como bootstrap físico mientras toda la semántica del sistema operativo vive en el runtime Alset.

Después se puede sustituir el bootstrap por una imagen AlsetOS dedicada sin cambiar el modelo de organismo.
