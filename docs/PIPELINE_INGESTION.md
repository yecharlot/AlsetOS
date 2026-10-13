# Pipeline de Ingestión Financiera

## Flujo completo

```
PDF / XLS / XLSX / DOC / DOCX
        ↓
   Extractores
        ↓
 JSON Canónico  alset.finance.statement/1.0
        ↓
  Traductor ORGES
        ↓
  draft .aos + Genes (ratios, TIR, VPN)
        ↓
  RootCID + modo Auditoría / Tokenización
```

## JSON Canónico (alset.finance.statement/1.0)

```json
{
  "schema": "alset.finance.statement/1.0",
  "entity": {
    "name": "Empresa Ejemplo SAS",
    "nit": "900123456",
    "period": "2025-Q3",
    "currency": "COP"
  },
  "balance": {
    "assets": {
      "current": 150000000,
      "non_current": 320000000,
      "total": 470000000
    },
    "liabilities": {
      "current": 80000000,
      "non_current": 120000000,
      "total": 200000000
    },
    "equity": 270000000
  },
  "income": {
    "revenue": 410000000,
    "cogs": 180000000,
    "gross": 230000000,
    "opex": 95000000,
    "ebitda": 135000000,
    "ebit": 120000000,
    "interest": 15000000,
    "tax": 28000000,
    "net": 77000000
  },
  "cashflow": {
    "operating": 110000000,
    "investing": -45000000,
    "financing": -20000000,
    "net": 45000000
  },
  "source": {
    "files": ["balance_2025q3.pdf", "dre_2025q3.xlsx"],
    "extracted_at": "2026-10-13T15:00:00Z"
  }
}
```

## Extractores recomendados (Go)

| Formato | Librería recomendada          | Notas                          |
|---------|-------------------------------|--------------------------------|
| PDF     | pdfcpu + regex / o llama a pdfplumber vía exec | Tablas DRE / Balance          |
| XLS/XLSX| github.com/xuri/excelize/v2  | Hojas con nombres normalizados |
| DOCX    | github.com/nguyenthenguyen/docx o similar | Tablas + texto                |

En la primera versión se puede aceptar JSON ya normalizado y CSV, y dejar PDF/XLS como “best effort” + warning.

## API

```
POST /api/ingest
Content-Type: application/json

{
  "kind": "finance-json" | "finance-pdf" | "finance-xls" | "finance-docx",
  "name": "analisis-acme-2025q3",
  "text": "<contenido o base64>",
  "payload": { ... json canónico ... }
}
```

Respuesta:
```json
{
  "ok": true,
  "draft_aos": "(flow analisis-acme-2025q3 ...)",
  "entities": ["balance", "dre", "flujo", "analisis-financiero"],
  "warnings": [],
  "mode": "ingest"
}
```

## Mapeo a ORGES

- `balance` → orge con state de activos/pasivos/patrimonio
- `dre` → orge con ingresos, costos, utilidad
- `flujo` → orge cashflow
- `analisis-financiero` → orge que requiere los anteriores + genes de ratios/TIR/VPN
