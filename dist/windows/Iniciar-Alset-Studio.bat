@echo off
cd /d "%~dp0"
title Alset Studio
echo.
echo  Alset Studio — Arquitecto + Editor + Preview
echo  http://127.0.0.1:5178/
echo.
echo  1) Pestana Arquitecto: entrevista y generar ORGES
echo  2) Pestana Editor: Aplicar flow, Preview, Desplegar
echo.
if not exist apps mkdir apps
start "" http://127.0.0.1:5178/
AlsetStudio.exe studio --addr 127.0.0.1:5178 --apps apps
pause
