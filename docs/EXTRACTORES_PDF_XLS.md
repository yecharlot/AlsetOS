# Extractores PDF / XLS / DOCX — Paso siguiente

## Archivos nuevos en este paquete

```
ingest/
├── types.go          ← FinanceStatement canónico
├── extract_xls.go    ← Excel (.xlsx / .xls) con excelize
├── extract_pdf.go    ← PDF (pdftotext + fallback de texto)
├── extract_docx.go   ← Word (.docx) sin dependencias externas
└── api.go            ← ExtractFromFile() unificado
```

## Dependencia

Solo para Excel:

```bash
go get github.com/xuri/excelize/v2
```

PDF usa `pdftotext` (poppler-utils) si está instalado; si no, hace un fallback de texto crudo (limitado).

DOCX no necesita dependencias (parsea el ZIP + XML internamente).

## Uso desde código

```go
import "tu.modulo/ingest"

res, err := ingest.ExtractFromFile("/ruta/estados_2025q3.xlsx")
if err != nil {
    // manejar
}
jsonBytes, _ := res.ToCanonicalJSON()
// → pasar a mind.IngestFinance con kind=finance-json
```

## Cómo cablearlo en AlsetOS (handler HTTP)

En `cmd/alset-plataforma` o donde esté el endpoint de ingestión / upload:

```go
// POST /api/ingest/upload  (multipart)
file, header, err := r.FormFile("file")
// guardar en tmp
tmp := filepath.Join(os.TempDir(), header.Filename)
// ...
res, err := ingest.ExtractFromFile(tmp)
if err != nil { ... }

// Convertir a mind.FinanceStatement (o unificar el tipo)
// luego:
draft := emitFinanceAOS(name, *convert(res.Statement))
```

O, más simple: el frontend sube el archivo, el backend extrae → JSON canónico → llama a `/api/ingest` con `kind=finance-json`.

## Calidad esperada

| Formato | Calidad típica | Notas |
|---------|----------------|-------|
| XLS/XLSX | Alta | Si las filas tienen etiquetas estándar (Activo corriente, Utilidad neta…) |
| DOCX    | Media-Alta     | Depende de que el texto esté en párrafos legibles |
| PDF texto | Media        | Mejor con `pdftotext`. PDFs escaneados necesitan OCR (no incluido) |
| PDF escaneado | Baja      | Requiere Tesseract/OCR externo |

## Recomendación de producción

1. **Primera línea**: Excel bien estructurado o JSON canónico.
2. **PDF**: pre-procesar con **pdfplumber** (Python) o Camelot y generar el JSON.
3. Este extractor Go sirve como “best-effort” dentro del nodo para no depender de Python.

## Ejemplo de prueba rápida

1. Crea un Excel con dos columnas: etiqueta | valor
2. Filas: `Activo corriente`, `150000000`, `Pasivo corriente`, `80000000`, `Utilidad neta`, `77000000`…
3. `ExtractFromFile("prueba.xlsx")` → debe rellenar balance e income.
4. Pasa el JSON resultante a `IngestFinance` → obtienes el `.aos` con genes.

## Próximo refinamiento posible

- Detectar automáticamente el periodo (2025-Q3, dic-2025…)
- Multi-año (columnas de años anteriores)
- OCR opcional vía Tesseract
- Plantillas por país (Colombia NIIF, US GAAP…)
