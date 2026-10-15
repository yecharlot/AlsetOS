#!/usr/bin/env bash
# Paquete portable Windows (amd64) — Studio + Finance CLI
# Objetivo: W10/W11. Go moderno ya no soporta oficialmente Windows 7.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
mkdir -p cmd/alset-plataforma/web
cp -f studio/index.html studio/studio.css studio/studio.js cmd/alset-plataforma/web/ 2>/dev/null || true
OUT="$ROOT/dist/windows"
mkdir -p "$OUT/samples/finance" "$OUT/docs"
export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64
# Toolchain: usa go.mod (1.26) vía auto si está disponible
export GOTOOLCHAIN="${GOTOOLCHAIN:-auto}"

echo "==> AlsetStudio.exe"
go build -trimpath -ldflags="-s -w" -o "$OUT/AlsetStudio.exe" ./cmd/alset-plataforma

echo "==> AlsetFinance.exe"
go build -trimpath -ldflags="-s -w" -o "$OUT/AlsetFinance.exe" ./cmd/alset-finance

cp -f ejemplos/finance/sample.json "$OUT/samples/finance/" 2>/dev/null || true
cp -f ejemplos/finance/analisis-financiero.aos "$OUT/samples/finance/" 2>/dev/null || true
cp -f plantillas/analisis-financiero.aos "$OUT/samples/finance/" 2>/dev/null || true
cp -f docs/GUIA_FINANCE.md docs/ROADMAP_SAAS.md docs/ARQUITECTO_MODOS_DE_TRABAJO.md "$OUT/docs/" 2>/dev/null || true

cat > "$OUT/Iniciar-Alset-Studio.bat" << 'BAT'
@echo off
cd /d "%~dp0"
title Alset Studio — Arquitecto ORGES
echo.
echo  Alset Studio + Finanzas + Mind
echo  http://127.0.0.1:5178/
echo.
if not exist apps mkdir apps
start "" http://127.0.0.1:5178/
AlsetStudio.exe studio --addr 127.0.0.1:5178 --apps apps
pause
BAT

cat > "$OUT/Analizar-Finanzas.bat" << 'BAT'
@echo off
cd /d "%~dp0"
title Alset Finance
echo Uso: arrastra un .json/.csv sobre este BAT o edita la ruta abajo.
set FILE=samples\finance\sample.json
if not "%~1"=="" set FILE=%~1
echo Analizando: %FILE%
AlsetFinance.exe -file "%FILE%" -out samples\finance\ultimo-draft.aos
echo.
echo Draft escrito en samples\finance\ultimo-draft.aos
echo Carga ese .aos en Alset Studio (editor) o modo Finanzas.
pause
BAT

cat > "$OUT/README.txt" << 'TXT'
Prism@.TEC / AlsetOS — paquete Windows portable
================================================
Requisitos: Windows 10 o Windows 11 (64-bit).
No necesita instalar Go ni npm.

*** Windows 7 ***
El toolchain Go actual (1.21+) ya no soporta Windows 7 de forma oficial.
Este build es para W10/W11 amd64.

Contenido
---------
AlsetStudio.exe     Studio + editor .aos + Mind + modos Arquitecto + deploy ANS local
AlsetFinance.exe    CLI: PDF/CSV/JSON/DOCX → ratios (liquidez, ROE, ROA, margen…) → draft .aos
Iniciar-Alset-Studio.bat
Analizar-Finanzas.bat
samples\finance\    JSON y plantillas de ejemplo
docs\               Guías

Arranque rápido
---------------
1. Doble clic en Iniciar-Alset-Studio.bat
2. Navegador → http://127.0.0.1:5178/
3. Panel Finanzas: "Cargar ejemplo" → "Analizar → draft ORGES"
4. O modo Análisis / Modelo / Ingestión / Implementación / Auditoría
5. Desplegar producción → apps\<nombre>\

CLI finanzas
------------
Analizar-Finanzas.bat
  o:
AlsetFinance.exe -file samples\finance\sample.json -out draft.aos

Formatos de entrada
-------------------
- JSON canónico alset.finance.statement/1.0 (recomendado, fiable)
- CSV etiqueta,valor
- PDF/DOCX best-effort (texto embebido; escaneados necesitan OCR externo)
- XLSX: convierte a JSON/CSV o usa excelize en builds con esa lib

Arquitectura del rol
--------------------
Arquitecto de negocios: describe dominio o alimenta estados financieros
→ ORGES (con o sin UI)
→ pruebas
→ deploy local (ANS apps/)
→ modo Auditoría/Tokenizar como puente a red PrismaTEC

SaaS
----
El mismo binario puede escucharse en 0.0.0.0 detrás de un reverse proxy.
Ver docs\ROADMAP_SAAS.md
TXT

echo "OK package:"
ls -lah "$OUT" | head -20
