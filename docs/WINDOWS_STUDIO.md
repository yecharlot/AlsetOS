# Alset Studio en Windows (sin Go ni npm)

Sí es posible: el Studio se empaqueta en **un solo `.exe`** con la UI embebida.

## Generar el portable (en una máquina con Go)

```bash
./scripts/build-windows.sh
# → dist/windows/AlsetStudio.exe
# → dist/windows/Iniciar-Alset-Studio.bat
```

Copia toda la carpeta `dist/windows/` a cualquier PC Windows 10/11.

## Uso en el PC destino

1. Ejecuta `Iniciar-Alset-Studio.bat`
2. Navegador → `http://127.0.0.1:5178/`
3. Apps desplegadas → carpeta `apps\` al lado del exe

## Qué incluye el exe

- Runtime LISPAI-AOS + API
- Studio (HTML/CSS/JS embebidos)
- Deploy multi-instancia a `apps/{nombre}/`
