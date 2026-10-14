# Studio del Arquitecto de Sistemas

El Studio no es solo un constructor de pantallas: es el **puesto de trabajo del Arquitecto** y del **tokenizador de modelos de negocio**.

## Barra de modos

| Modo | Uso |
|------|-----|
| **1 · Análisis** | Entrevista / dominio → roadmap de digitalización |
| **2 · Ingestión** | Schema, SQL, JSON, PDF/XLS financiero → draft ORGES |
| **3 · Modelo ORGES** | Reglas y organismos de dominio (UI opcional) |
| **4 · Implementación** | Surface UI + preview + deploy |
| **5 · Auditoría / Tokenizar** | Políticas, Pulse, fingerprint del manifiesto |

## Entrevista técnica → ORGES

1. Completa empresa, sector, procesos, requisitos funcionales y no funcionales.
2. **Generar ORGES** — Mind (`POST /api/work/domain`) + generador local de respaldo.
3. El `.aos` se carga en el **editor inteligente** (coloreado, Live, sugerencias).
4. **Solo análisis / roadmap** fuerza modo Análisis.

## Reglas manuales → ORGES

1. Escribe políticas en lenguaje natural o viñetas.
2. **Transformar en ORGES** — Mind + heurísticas (stock, ventas, roles).
3. **Enviar al editor .aos** para afinar a mano con el agente Mind del editor.

## APIs

- `GET /api/work/modes`
- `POST /api/work/domain` — `{ message, work_mode, name, source }`
- `POST /api/ingest` — finanzas / schema

## Arranque

```bash
go run ./cmd/alset-plataforma studio --addr 127.0.0.1:5178
# http://127.0.0.1:5178/
```
