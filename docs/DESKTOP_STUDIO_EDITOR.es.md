# Alset Desktop — gestor de ventanas real

## Probar ya

```bash
cd AlsetOS
export GOTOOLCHAIN=go1.26.0
./scripts/run-alset-desktop.sh
# o:
go run ./cmd/alset-desktop-bridge -addr 127.0.0.1:7420 \
  -shell desktop/shell \
  -web ../Alset-LISPAI-Runtime/web \
  -alsetos ./bin/alsetos
```

Abre **http://127.0.0.1:7420/**

## Qué incluye

| Capacidad | Comportamiento |
|-----------|----------------|
| **Iconos** | Arrastrar y soltar; posición en `localStorage` |
| **Ventanas** | Mover (barra título), redimensionar, minimizar, maximizar, cerrar |
| **Studio / Editor** | Se abren **dentro** del escritorio (iframe), no en otra pestaña |
| **Terminal** | LispAI + alsetState + mind/zyrion/neural |
| **Mind · Zyrion · Neural · Silogismos** | Servicios del sistema en `/v1/...` |

## APIs cognitivas (OS + apps)

| Ruta | Uso |
|------|-----|
| `POST /v1/mind/tick` | `{ "text": "..." }` → decisión Mind |
| `POST /v1/zyrion` | `{ "a", "b" }` → valor ternario |
| `POST /v1/syllogism/assert` | hecho `{ s, r, o, c }` |
| `POST /v1/syllogism/infer` | cadena de hechos |
| `POST /v1/syllogism/ask` | consulta |
| `GET/POST /v1/neural` | pesos 0..1 |

Las apps creadas en Studio/Editor pueden llamar las **mismas** rutas: coexisten con el sistema operativo.

## Iconos de escritorio

Doble clic abre la app. Arrastra para reposicionar (se guarda solo).
