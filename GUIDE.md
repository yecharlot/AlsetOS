# GUIDE.md — Guía obligatoria AlsetOS (consulta primero)

**Repositorio:** https://github.com/yecharlot/AlsetOS  
**Actualización:** 2026-10-15  
**Audiencia:** desarrolladores e IAs que continúan el trabajo.

Este archivo **unifica** la orientación del proyecto. Los `.md` temáticos en `docs/` siguen siendo referencia profunda; **empieza aquí**.

---

## 1. Visión en una frase

**AlsetOS** modela negocios como **organismos digitales (ORGES)** declarados en **LISPAI-AOS (`.aos`)**: estado, capacidades, reglas ARL, contratos `bind`, tests y deploy opcional.  
**PrismaTEC-Core** aporta identidad, Pulse, CID y red de evidencia.  
**Tokenizar** un negocio ≠ coin especulativo: es **manifiesto versionado + huella CID + cadena append-only de eventos**.

---

## 2. Mapa del monorepo (qué tocar)

| Ruta | Rol |
|------|-----|
| `cmd/alset-plataforma` | Studio HTTP + APIs + deploy apps |
| `cmd/alset-finance` | CLI estados financieros → ratios → draft `.aos` |
| `studio/` | UI Arquitecto / editor / preview (se embebe en `cmd/.../web`) |
| `declarativo/` | Runtime ORGES, ARL, codegen |
| `lispai/` | Parser S-expr AOS |
| `mind/` | Agente código, modos, entrevista, finanzas |
| `gene/finance.go` | Razones financieras + VPN/TIR + Zyrion |
| `ingest/` | PDF/CSV/JSON/DOCX → statement canónico |
| `token/` | **Tokenización de modelos** (CID + chain local) |
| `dist/windows/` | Ejecutables portables W10/W11 |
| `docs/` | Guías temáticas (LISPAI, finance, desktop, etc.) |

---

## 3. Producto SaaS del Arquitecto (flujo comercial)

1. **Entrevista técnica** (empresa, sector, procesos, RF/RNF, sistemas, dolores).  
2. **Informe automático:** resumen, estructura, mejoras, riesgos, plan ORGES.  
3. **Draft `.aos`** (modo *modelo* sin UI, o *implementación* con surface).  
4. **Tokenizar:** `POST /api/token/mint` → CID del manifiesto + eslabón `mint`.  
5. Opcional: finanzas (JSON/CSV/PDF) → genes → ORGES de análisis.  
6. Opcional: deploy local `/apps/{nombre}/` o sombra PrismaTEC.

APIs clave:

- `POST /api/architect/interview`  
- `POST /api/token/mint` · `GET /api/token/list` · `POST /api/token/append`  
- `POST /api/finance/upload` · `POST /api/ingest`  
- `POST /api/flow/apply|test|deploy|action`  
- `POST /api/mind/analyze` · `POST /api/work/domain`

---

## 4. Arranque rápido

```bash
# Studio
go run ./cmd/alset-plataforma studio --addr 127.0.0.1:5178 --apps ./apps

# Finanzas CLI
go run ./cmd/alset-finance -file ejemplos/finance/sample.json -out /tmp/f.aos

# Tests núcleo studio-related
go test ./token/ ./gene/ ./ingest/ ./mind/ ./declarativo/ -count=1
```

Windows: `dist/windows/Iniciar-Alset-Studio.bat` (requiere **W10/W11 x64**; Go ya no soporta W7).

---

## 5. Historial de trabajo (por fecha)

### 2026-10-15 (UX responsiva)
- Layout a pantalla completa: pestañas **Arquitecto** vs **Editor·Preview·Deploy**; modos muestran/ocultan tarjetas; panel Resultado visible.

### 2026-10-15
- Paquete `token/`: mint de modelos de negocio, CID `cid1:sha256`, cadena de eventos, persistencia `data/tokens`.
- `mind/architect.go`: informe completo de entrevista (estructura, mejoras, riesgos, plan ORGES, oferta comercial).
- APIs `/api/architect/interview`, `/api/token/*`.
- Studio: campo dolores, informe Arquitecto, botón tokenizar.
- Finanzas: upload multipart, extractor CSV, panel Finanzas, `AlsetFinance.exe` en `dist/windows`.
- Este **GUIDE.md** unificado.

### 2026-10-14
- Modos Arquitecto (análisis / ingestión / modelo / implementación / auditoría).
- Vertical financiero: genes, ingest, plantilla, CLI `alset-finance`.
- Documentación `docs/GUIA_FINANCE.md`, `ROADMAP_SAAS.md`, `ARQUITECTO_*`.

### 2026-10-12 — 2026-10-13
- Studio ORGES: multi-instancia apps, preview≈deploy, props UI, Inspector Mind.
- Tabs/drawer reales, API action de apps, portable `AlsetStudio.exe`.
- Manuales PDF creación de ORGES / dominio→ORGES / LISPAI-AOS.

### 2026-10-11
- LISPAI-AOS flows, `alset-plataforma flow|studio`.
- Reglas primero, ciclo organismo, ejemplo pedidos, UiRuntime / desktop (línea paralela).

---

## 6. Pendientes (priorizados)

| Prioridad | Ítem | Notas |
|-----------|------|--------|
| P0 | Excel XLSX nativo (excelize) | Hoy JSON/CSV; PDF best-effort |
| P0 | Sincronizar `studio/` → `web/` en cada release | Script `build-windows.sh` ya copia |
| P1 | Bridge opcional a PrismaTec-Core (Pulse hub / AIP) | Mismo concepto de token; proceso o lib |
| P1 | Firma Ed25519 de eslabones de token | Alineado a identidad de nodo Core |
| P1 | Piloto SaaS: un dominio real vendido (solo informe+modelo) | Validar negocio del Arquitecto |
| P2 | OCR para PDF escaneados | Fuera de banda → JSON |
| P2 | Multi-nodo P2P para replicar registry de tokens | Usar `cmd/alset-p2p` / Core network |
| P2 | Unificar más `docs/*.md` obsoletos en anexos | Mantener GUIDE como índice |
| P3 | Desktop TinyCore ISO | Documentado; hardware real |

---

## 7. Ideas de producto más potentes (SaaS)

1. **Arquitecto as a Service:** entrevista + informe + `.aos` + token CID en 48 h (sin UI).  
2. **Auditoría de divergencia:** sombra del sistema viejo vs estado ORGES (Core Pulse).  
3. **Plantillas verticales de pago:** comercio, finanzas, logística (ÁbacoPhy como referencia).  
4. **Marketplace de genes** (ratios, stock, rutas) versionados por CID.  
5. **Compliance pack:** export de cadena de evidencia para dueños que temen manipulación de datos.

---

## 8. Relación con otros repos

| Repo | Uso |
|------|-----|
| [PrismaTec-Core](https://github.com/yecharlot/PrismaTec-Core) | Organism, Pulse, CID store, red, AIP, Studio control |
| [PrismaTec](https://github.com/yecharlot/PrismaTec) | Docs/análisis; código de nodo migrado hacia Core |
| [AbacoPhy](https://github.com/yecharlot/AbacoPhy) | Vertical contable PyME |
| [Alset-LISPAI-Runtime](https://github.com/yecharlot/Alset-LISPAI-Runtime) | Studio web LispAI histórico |
| [Alset-JS-Runtime](https://github.com/yecharlot/Alset-JS-Runtime) | UI por pulsos |

---

## 9. Convención para PRs / IAs

1. Leer **este GUIDE.md**.  
2. No crear docs sueltos sin enlazarlos aquí (fecha + pendiente).  
3. Tras cambiar `studio/*`, copiar a `cmd/alset-plataforma/web/`.  
4. Regenerar `dist/windows` con `./scripts/build-windows.sh` si afecta al portable.  
5. Preferir **un vertical cerrado** a nuevas abstracciones globales.

---

## 10. Índice rápido de docs temáticos

- `docs/LISPAI_AOS.md` · `docs/ATS001_LISPAI_AOS.md`  
- `docs/GUIA_FINANCE.md` · `docs/GENES_FINANCIEROS.md` · `docs/PIPELINE_INGESTION.md`  
- `docs/ARQUITECTO_MODOS_DE_TRABAJO.md` · `docs/ARQUITECTO_STUDIO.md`  
- `docs/ROADMAP_SAAS.md` · `docs/STATUS.md`  
- `docs/PACKAGING_WINDOWS.md` · `docs/ALSET_DESKTOP.es.md`  

**Fin de la guía obligatoria.**
