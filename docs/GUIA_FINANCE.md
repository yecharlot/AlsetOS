# Guía práctica — Vertical financiero AlsetOS

**Base:** commit en `main` con genes, ingestión y CLI.  
**Objetivo:** de un archivo de estados financieros a ratios (Zyrion) + draft `.aos` de ORGES.

## Qué hace este vertical

1. **Lee** PDF / XLS / DOCX / JSON de balance, estado de resultados y flujo.
2. **Normaliza** a `alset.finance.statement/1.0` (JSON canónico).
3. **Calcula genes:** liquidez, prueba ácida, endeudamiento, ROE, ROA, margen, VPN, TIR, alerta.
4. **Emite** un flow `.aos` con organismos `balance`, `dre`, `flujo`, `analisis-financiero`.
5. **Mind** (`POST /api/ingest` con `kind: finance-json`) genera el mismo draft desde Studio.

## Requisitos

- Go 1.22+ (el módulo declara 1.26; en local puedes usar `GOTOOLCHAIN=local` si hace falta).
- Opcional Excel: dependencia `github.com/xuri/excelize/v2` (ya en `go.mod` si corriste `go mod tidy`).
- PDF: mejor con `pdftotext` (poppler-utils). Sin él hay fallback limitado.
- DOCX: sin dependencias extra.

## Linux / macOS — prueba en 2 minutos

```bash
git pull origin main
cd AlsetOS

# Ratios + preview .aos desde JSON de ejemplo
go run ./cmd/alset-finance -file ejemplos/finance/sample.json

# Guardar el flow completo
go run ./cmd/alset-finance -file ejemplos/finance/sample.json -out /tmp/demo-finance.aos

# Tests de genes
go test ./gene/ -count=1 -run Finance -v
go test ./gene/ -count=1
```

### Studio / API (red local)

```bash
go run ./cmd/alset-plataforma studio --addr 127.0.0.1:5178
```

Luego:

```bash
curl -s -X POST http://127.0.0.1:5178/api/ingest \
  -H "Content-Type: application/json" \
  -d @- <<'EOF'
{
  "kind": "finance-json",
  "name": "demo-pyme",
  "payload": $(cat ejemplos/finance/sample.json)
}
EOF
```

(En Windows PowerShell conviene cargar el JSON en una variable y enviarlo.)

Modos del Arquitecto: **Ingestión** acepta ahora `finance-json`, `finance-xls`, `finance-pdf`, `finance-docx`.

## Windows — nodo financiero

1. Clona o descarga el repo.
2. Compila (en una máquina con Go) o usa el portable de Studio ya en `dist/`:

```bat
go build -o alset-finance.exe ./cmd/alset-finance
go build -o alset-plataforma.exe ./cmd/alset-plataforma
```

3. Ejecuta `scripts\start-finance-node.bat` (ajusta rutas si los `.exe` están en otra carpeta).
4. Abre el navegador en la URL del Studio y usa modo Ingestión o el CLI:

```bat
alset-finance.exe -file ejemplos\finance\sample.json -out demo.aos
```

Detalle de empaquetado: `docs/PACKAGING_WINDOWS.md` y `docs/WINDOWS_STUDIO.md`.

## Interpretar la salida de ratios

| Zyrion | Significado |
|--------|------------|
| **SI (1)** | Dentro de umbral sano |
| **NO (0)** | Fuera de umbral (alerta) |
| **INCIERTO (2)** | Dato faltante o zona gris |

Ejemplo: liquidez &lt; 1.0 → NO; endeudamiento &gt; 0.70 → NO.

## Plantilla lista

`plantillas/analisis-financiero.aos` — ORGES de ejemplo con números ilustrativos.

## Documentación relacionada

- `docs/PIPELINE_INGESTION.md` — pipeline completo
- `docs/GENES_FINANCIEROS.md` — fórmulas
- `docs/EXTRACTORES_PDF_XLS.md` — extractores
- `docs/ROADMAP_SAAS.md` — multi-tenant / tokenización (fase posterior)

## Qué no promete esta rebanada

- OCR de PDF escaneados (hace falta herramienta externa).
- SaaS multi-tenant en producción (roadmap, no incluido como servicio desplegado).
- Que el Studio Windows ya empaquete solo el vertical financiero sin compilar.

La rebanada **sí** te deja probar localmente: archivo → ratios → `.aos` → ingest Mind.


## Panel visual LISPAI-AOS (UI moderna)

El flow generado incluye el organismo **`ui-finance`** (`kind UI`):

- Splash con entidad y periodo
- Gradient animado de marca
- Badges reactivos de ratios (liquidez, ROE, ROA, margen, endeudamiento)
- Tabs: Resumen · Balance · DRE · Flujo · Alertas
- Botones de acción (`finance.recalcular`, `finance.exportar`, marcar revisado)
- Surface: `ui-finance` — listo para **Aplicar flow** y **Desplegar** en Studio

### Probar la UI en Studio

```bash
go run ./cmd/alset-plataforma studio --addr 127.0.0.1:5178
```

1. Abre **Ejemplos** → `analisis-financiero` (o carga `plantillas/analisis-financiero.aos`).
2. **Aplicar flow** — el preview muestra el panel.
3. Cambia de pestaña (Resumen / Balance / Alertas).
4. **Desplegar producción** → `/apps/analisis-financiero/` sin header del Studio.

### Generar .aos con UI desde datos

```bash
go run ./cmd/alset-finance -file ejemplos/finance/sample.json -out /tmp/demo-ui.aos
# El draft incluye ui-finance con ratios precalculados
```

Los ratios se precalculan al emitir el draft para que el panel muestre números reales de inmediato; los genes siguen declarados en el organismo de análisis.
