# Salto: deploy → escritorio, persistencia, un comando

## Un comando

```bash
cd AlsetOS
export GOTOOLCHAIN=go1.26.0
./scripts/run-alset-desktop.sh
```

http://127.0.0.1:7420/

## Deploy desde Studio (dentro del SO)

1. Abre **Studio** (icono o menú).
2. Arma la UI → **Desplegar**.
3. El SO recibe `postMessage`, crea **icono** y **ventana** con la app en `/apps/<nombre>/`.
4. Reinicia el navegador: iconos de apps instaladas y ventanas de apps se restauran.

## Persistencia

| Qué | Dónde |
|-----|--------|
| Iconos | `localStorage` |
| Tema | `localStorage` |
| Ventanas de apps | `localStorage` |
| Binarios de apps | `~/.alset-desktop/apps/` |

## ISO / payload

```bash
./scripts/package-alset-desktop-iso.sh
```
