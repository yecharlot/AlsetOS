@echo off
REM Nodo Alset financiero — Windows (un clic si los exe estan junto a este script)
setlocal
cd /d "%~dp0\.."

if exist "alset-plataforma.exe" (
  start "Alset Studio" alset-plataforma.exe studio --addr 127.0.0.1:5178
) else if exist "dist\alset-studio-windows\AlsetStudio.exe" (
  start "Alset Studio" dist\alset-studio-windows\AlsetStudio.exe
) else (
  echo Compila primero: go build -o alset-plataforma.exe ./cmd/alset-plataforma
  echo O usa el portable en dist\alset-studio-windows\
)

if exist "alset-finance.exe" (
  echo.
  echo Ejemplo CLI:
  echo   alset-finance.exe -file ejemplos\finance\sample.json
) else (
  echo Compila CLI: go build -o alset-finance.exe ./cmd/alset-finance
)

timeout /t 2 >nul
start http://127.0.0.1:5178/
echo Studio en http://127.0.0.1:5178/
endlocal
