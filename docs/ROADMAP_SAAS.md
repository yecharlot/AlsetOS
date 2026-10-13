# Roadmap SaaS Vertical — Alset as a Service

## Visión

Convertir el ecosistema Alset en una plataforma donde un Arquitecto puede:

1. Ingerir documentos financieros / contables / de negocio
2. Generar ORGES vivos
3. Analizar con genes (ratios, TIR, VPN, etc.)
4. Tokenizar el modelo (RootCID + Pulse + evidencia)
5. Desplegar en red descentralizada o como SaaS multi-tenant
6. Ofrecer verticales (Finanzas, Contabilidad ÁbacoPhy, Ventas Latati, etc.)

## Fases

### Fase 0 — Base (HECHA — 2026-10-13)
- AlsetOS con modos Arquitecto (Análisis, Modelo, Ingestión, Implementación, Auditoría)
- Studio + editor .aos + Mind Agent
- Runtime ORGES + LispAI + Zyrion + P2P + autonomía
- Portable Windows de Studio
- ÁbacoPhy (partida doble + tenants)
- PrismaTec-Core (organismos, AIP, Pulse)

### Fase 1 — Vertical Financiero (1-2 semanas)
- [ ] Integrar `ingest_finance.go` + `gene/finance.go`
- [ ] Extender Studio con kind finance-*
- [ ] Plantillas .aos de análisis
- [ ] Empaquetado Windows de un solo clic
- [ ] Pruebas con estados financieros reales (JSON canónico primero)

### Fase 2 — SaaS multi-tenant (2-4 semanas)
- [ ] Nodo multi-tenant (patrón ÁbacoPhy + Latati)
- [ ] ANS por cliente (`cliente.finance.ans`)
- [ ] Dashboard de alertas Zyrion
- [ ] API pública + autenticación
- [ ] Tokenización automática del RootCID del modelo
- [ ] Billing / planes (opcional)

### Fase 3 — Ecosistema
- [ ] Marketplace de Genes
- [ ] Studio multi-usuario en la nube
- [ ] Red de nodos intercambiando organismos tokenizados
- [ ] Contratos digitales Zyrion ejecutables
- [ ] Más verticales (legal, logística, salud, educación…)

## Arquitectura del Arquitecto

```
Usuario (Arquitecto)
    ↓
Alset Studio (modos)
    ↓
ORGES + Genes financieros
    ↓
RootCID + Pulse firmado
    ↓
Nodo Alset (local Windows / cloud)
    ↓
Red libp2p + ANS + PrismaTec-Core
    ↓
Tokenización + SaaS multi-tenant
```

## Monetización posible

- SaaS por tenant / por análisis
- Genes premium
- Tokenización y anclaje de modelos de negocio
- Consultoría de Arquitecto Alset (tú)

## Próximo hito concreto

1. Integrar los archivos de este paquete en AlsetOS
2. Probar ingestión con un JSON de ejemplo
3. Generar el primer ORGE de análisis financiero
4. Empaquetar el portable Windows
