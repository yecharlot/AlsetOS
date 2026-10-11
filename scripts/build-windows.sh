#!/usr/bin/env bash
# Genera ejecutable Windows portable del Alset Studio ORGES (sin Go/npm en el PC destino).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
# sincronizar UI embebida
mkdir -p cmd/alset-plataforma/web
cp -f studio/index.html studio/studio.css studio/studio.js cmd/alset-plataforma/web/
OUT="$ROOT/dist/windows"
mkdir -p "$OUT"
export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64
go build -trimpath -ldflags="-s -w" -o "$OUT/AlsetStudio.exe" ./cmd/alset-plataforma
# launcher
cat > "$OUT/Iniciar-Alset-Studio.bat" << 'BAT'
@echo off
cd /d "%~dp0"
echo Alset Studio ORGES
echo Abre http://127.0.0.1:5178/
start "" http://127.0.0.1:5178/
AlsetStudio.exe studio --addr 127.0.0.1:5178 --apps apps
pause
BAT
cat > "$OUT/README.txt" << 'TXT'
Alset Studio ORGES — portable Windows 10/11
===========================================
No requiere instalar Go ni npm.

1. Doble clic en Iniciar-Alset-Studio.bat
2. Se abre el navegador en http://127.0.0.1:5178/
3. Crea ORGES, prueba y Desplegar produccion.
4. Las apps quedan en la carpeta apps\ junto al .exe

Solo ejecutable: AlsetStudio.exe studio --addr 127.0.0.1:5178 --apps apps
TXT
echo "OK: $OUT/AlsetStudio.exe"
ls -la "$OUT"
