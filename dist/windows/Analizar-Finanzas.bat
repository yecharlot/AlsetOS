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
