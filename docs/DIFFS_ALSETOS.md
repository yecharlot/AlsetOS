# Diffs exactos para integrar en AlsetOS

## 1. mind/modes.go

En la función `CatalogModes()`, reemplazar el bloque del modo Ingestión por:

```go
{
    ID: string(ModeIngest), Label: "Ingestión",
    Summary: "Importa schema, SQL, API REST o documentos financieros (PDF/XLS/DOCX/JSON) y genera draft ORGES + genes de análisis.",
    RequiresUI: false,
    Outputs:    []string{"IR intermedio", "draft .aos", "ratios + TIR/VPN", "warnings"},
    NextSteps:  []string{"Revisar en Modelo", "Firmar manifiesto", "Auditoría / Tokenización"},
},
```

## 2. mind/agent.go o el switch de Ingest

En `func (a *Agent) Ingest(req IngestRequest) IngestResponse` añadir al inicio del switch de `kind`:

```go
case "finance-json", "finance", "finance-pdf", "finance-xls", "finance-docx":
    return a.IngestFinance(req)
```

## 3. Nuevos archivos a copiar

- `mind/ingest_finance.go`  →  AlsetOS/mind/ingest_finance.go
- `gene/finance.go`         →  AlsetOS/gene/finance.go

## 4. go.mod

No se requieren dependencias nuevas para la versión JSON-first.
Cuando se implementen extractores PDF/XLS:

```
github.com/xuri/excelize/v2
```

(y opcionalmente un binding a pdfplumber o pdfcpu).

## 5. Studio

Ver `studio/patches.md`.

## 6. Prueba rápida

1. Arrancar Studio / plataforma.
2. Modo Ingestión → kind = finance-json.
3. Pegar el JSON de ejemplo de `docs/PIPELINE_INGESTION.md`.
4. Pulsar “Ingerir → draft .aos”.
5. Debe aparecer el flow con balance, dre, flujo y analisis-financiero + genes.
