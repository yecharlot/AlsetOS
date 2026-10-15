Prism@.TEC / AlsetOS — paquete Windows portable
================================================
Requisitos: Windows 10 o Windows 11 (64-bit).
No necesita instalar Go ni npm.

*** Windows 7 ***
El toolchain Go actual (1.21+) ya no soporta Windows 7 de forma oficial.
Este build es para W10/W11 amd64.

Contenido
---------
AlsetStudio.exe     Studio + editor .aos + Mind + modos Arquitecto + deploy ANS local
AlsetFinance.exe    CLI: PDF/CSV/JSON/DOCX → ratios (liquidez, ROE, ROA, margen…) → draft .aos
Iniciar-Alset-Studio.bat
Analizar-Finanzas.bat
samples\finance\    JSON y plantillas de ejemplo
docs\               Guías

Arranque rápido
---------------
1. Doble clic en Iniciar-Alset-Studio.bat
2. Navegador → http://127.0.0.1:5178/
3. Panel Finanzas: "Cargar ejemplo" → "Analizar → draft ORGES"
4. O modo Análisis / Modelo / Ingestión / Implementación / Auditoría
5. Desplegar producción → apps\<nombre>\

CLI finanzas
------------
Analizar-Finanzas.bat
  o:
AlsetFinance.exe -file samples\finance\sample.json -out draft.aos

Formatos de entrada
-------------------
- JSON canónico alset.finance.statement/1.0 (recomendado, fiable)
- CSV etiqueta,valor
- PDF/DOCX best-effort (texto embebido; escaneados necesitan OCR externo)
- XLSX: convierte a JSON/CSV o usa excelize en builds con esa lib

Arquitectura del rol
--------------------
Arquitecto de negocios: describe dominio o alimenta estados financieros
→ ORGES (con o sin UI)
→ pruebas
→ deploy local (ANS apps/)
→ modo Auditoría/Tokenizar como puente a red PrismaTEC

SaaS
----
El mismo binario puede escucharse en 0.0.0.0 detrás de un reverse proxy.
Ver docs\ROADMAP_SAAS.md
