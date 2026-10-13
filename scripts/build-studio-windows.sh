#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
OUT="$ROOT/dist/alset-studio-windows"
rm -rf "$OUT"
mkdir -p "$OUT/studio" "$OUT/apps"

export CGO_ENABLED=0 GOOS=windows GOARCH=amd64
go build -trimpath -ldflags="-s -w" -o "$OUT/AlsetStudio.exe" ./cmd/alset-plataforma

cp -a studio/. "$OUT/studio/"

cat > "$OUT/Iniciar-Alset-Studio.bat" << 'BAT'
@echo off
cd /d "%~dp0"
set ALSET_STUDIO=1
echo Alset Studio ORGES
echo http://127.0.0.1:5178/
start "" http://127.0.0.1:5178/
AlsetStudio.exe studio --addr 127.0.0.1:5178 --studio ./studio --apps ./apps
pause
BAT

cat > "$OUT/README.txt" << 'TXT'
Alset Studio (Windows portable)
1. Doble clic en Iniciar-Alset-Studio.bat
2. Abre http://127.0.0.1:5178/
No requiere Go ni npm instalados.
TXT

echo "OK $OUT"
ls -la "$OUT"
