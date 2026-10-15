# Modos de trabajo — Arquitecto y tokenizador de negocios

Alset Studio ya no asume que todo termina en pantallas. El cliente puede pedir solo
**análisis y ruta de digitalización**, un **modelo ORGES** ejecutable, **ingesta** desde
sistemas legados, **implementación** con UI, o **auditoría / anclaje** con PrismaTEC.

## Modos

| Modo | ¿UI obligatoria? | Salida principal |
|------|------------------|------------------|
| **Análisis** | No | Roadmap de digitalización + draft dominio |
| **Modelo** | No | `.aos` con ORGES, rules, binds, tests |
| **Ingestión** | No | Draft desde SQL / schema JSON / API REST |
| **Implementación** | Sí (preview/deploy) | App multi-dispositivo |
| **Auditoría** | No | Política shadow, Pulse, tokenización del manifiesto |

## APIs

- `GET /api/work/modes`
- `POST /api/work/domain` — `{ message, work_mode, name, source }`
- `POST /api/ingest` — `{ kind: sql|schema|rest, text|payload, name }`

## Repos que alimentan el objetivo

- **AlsetOS** — editor, runtime ORGES, modos, deploy
- **PrismaTec-Core** — organism, pulse, red de evidencia
- **PrismaTec** — demos / red
- **ÁbacoPhy / verticales** — plantillas de dominio comercial

El arquitecto firma el manifiesto; el token no es especulación: es **versión del modelo + cadena de evidencia**.

## UI del Studio (Arquitecto)
Ver [ARQUITECTO_STUDIO.md](ARQUITECTO_STUDIO.md): barra de modos, entrevista técnica y reglas → ORGES.
