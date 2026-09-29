# Arrancar Alset Desktop en hardware real (Tiny Core)

## Qué necesitas

1. ISO TinyCore (~25 MB) — base mínima.
2. Paquete `alset-desktop-0.1-linux-amd64.tgz` (bridge + shell + alsetos).
3. USB para ISO y disco/USB de datos con `tce=` para persistencia.

## Pasos

### 1. Grabar ISO
```bash
sudo dd if=TinyCore-current.iso of=/dev/sdX bs=4M status=progress && sync
```

### 2. Arrancar
BIOS → USB. Elige **TinyCore** (gráfico FLWM).

### 3. Instalar Alset
```sh
sudo mkdir -p /opt/alset
sudo tar xzf alset-desktop-0.1-linux-amd64.tgz -C /opt/alset
sudo chmod +x /opt/alset/bin/* /opt/alset/boot/alset-desktop.sh
```

### 4. Extensiones
```sh
tce-load -wi Xorg-7.7 flwm
# + navegador del mirror si existe
```

### 5. Lanzar
```sh
/opt/alset/boot/alset-desktop.sh
# UI: http://127.0.0.1:7420/
```

### 6. Persistencia
Boot: `tce=/mnt/<disco>/tce`  
Copia `/opt/alset` al disco de datos y engancha `bootlocal.sh`.

## Programar desde allí

0.1 = ejecución y control (organismos, shell).  
Compila Go en host (`GOOS=linux GOARCH=amd64 CGO_ENABLED=0`) y copia binarios a `/opt/alset/bin`.

## Recuperación
FLWM + terminal siempre disponibles si la shell falla.
