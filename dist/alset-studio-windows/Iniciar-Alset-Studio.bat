@echo off
cd /d "%~dp0"
echo Alset Studio ORGES
echo http://127.0.0.1:5178/
start "" http://127.0.0.1:5178/
AlsetStudio.exe studio --addr 127.0.0.1:5178 --studio ./studio --apps ./apps
pause
