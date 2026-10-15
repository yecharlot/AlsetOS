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
