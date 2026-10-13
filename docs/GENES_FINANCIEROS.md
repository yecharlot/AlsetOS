# Genes Financieros — Razones + VPN + TIR

Implementados en `gene/finance.go` y ejecutables desde ORGES / LispAI / Zyrion.

## Lista de genes

| ID                        | Entrada                          | Salida                          | Zyrion |
|---------------------------|----------------------------------|---------------------------------|--------|
| gene.ratio.liquidez       | activo_corriente, pasivo_corriente | valor float64                  | 0/1/2 |
| gene.ratio.prueba_acida   | ac, inventario, pc               | valor                           | 0/1/2 |
| gene.ratio.endeudamiento  | pasivo_total, activo_total       | valor                           | 0/1/2 |
| gene.ratio.roe            | utilidad_neta, patrimonio        | valor                           | 0/1/2 |
| gene.ratio.roa            | utilidad_neta, activo_total      | valor                           | 0/1/2 |
| gene.ratio.margen_neto    | utilidad_neta, ingresos          | valor                           | 0/1/2 |
| gene.vpn                  | flujos[], tasa, inversion_inicial | valor presente neto            | 0/1/2 |
| gene.tir                  | flujos[] (incluye inversión)     | tasa interna de retorno         | 0/1/2 |
| gene.alerta_financiera    | mapa de ratios + umbrales        | alerta bool + razones           | 0/1/2 |

## Umbrales por defecto (ajustables por dominio)

- Liquidez corriente < 1.0 → alerta
- Prueba ácida < 0.8 → alerta
- Endeudamiento > 0.7 → alerta
- ROE < 0.05 → alerta
- TIR < tasa de descuento del proyecto → alerta

## Uso en .aos

```lisp
(orge analisis-financiero
  (kind ORGES)
  (cap "finance.analizar" "finance.alertar")
  (requires "balance.entity" "1.0" "dre.entity" "1.0")
  (state
    (analisis.liquidez 0)
    (analisis.tir 0)
    (analisis.vpn 0)
    (analisis.alerta false))
  (gene gene.ratio.liquidez)
  (gene gene.tir)
  (gene gene.vpn)
  (rule liquidez-critica
    (when (lt analisis.liquidez 1.0))
    (then analisis.alerta true)))
```

## Implementación

Ver `gene/finance.go`. Las funciones son puras (sin side-effects) y devuelven valor + código Zyrion (0 = no, 1 = sí, 2 = incierto).
