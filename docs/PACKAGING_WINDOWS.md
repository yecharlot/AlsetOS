# Empaquetado Windows — Nodo Alset Financiero en 5 minutos

## Objetivo

Un zip/portable que el arquitecto abre y en menos de 5 minutos tiene:

- Alset Studio (con modos Arquitecto + ingestión financiera)
- Nodo Alset local
- Genes financieros
- Plantillas .aos
- Capacidad de tokenizar

## Estructura del portable

```
Alset-Finance-Node-Windows/
├── AlsetStudio.exe          ← ya existe en AlsetOS/dist/
├── alset-node.exe
├── alset-autonomia.exe
├── alset-p2p.exe
├── genes/
│   └── finance/             ← binarios o .go compilados
├── plantillas/
│   └── analisis-financiero.aos
├── data/                    ← persistencia local + CID
├── start-finance-node.bat
└── README-WINDOWS.txt
```

## start-finance-node.bat

```bat
@echo off
title Alset Finance Node
echo Iniciando Alset Finance Node...
start "" "AlsetStudio.exe"
timeout /t 2 >nul
start "" "alset-node.exe" --profile finance --port 8080
timeout /t 1 >nul
start http://127.0.0.1:7420
echo Studio y Nodo lanzados.
echo Studio: http://127.0.0.1:7420
echo Nodo:   http://127.0.0.1:8080
pause
```

## Cómo construirlo

1. Desde AlsetOS (con Go 1.26):
   ```
   go build -o dist/windows/alset-node.exe ./cmd/alset-node
   go build -o dist/windows/alset-autonomia.exe ./cmd/alset-autonomia
   go build -o dist/windows/alset-p2p.exe ./cmd/alset-p2p
   ```
2. Copiar el AlsetStudio.exe portable ya existente.
3. Incluir este paquete (`plantillas/`, `gene/`, docs).
4. Empaquetar en zip o con Inno Setup / 7-Zip SFX.

## Requisitos del usuario final

- Windows 10/11
- Ningún Go, npm ni dependencias instaladas (todo estático)
