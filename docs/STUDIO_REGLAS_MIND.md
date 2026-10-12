# Studio: reglas manuales + Mind

## Editor inteligente (.aos)
- Coloreado de palabras clave LISPAI-AOS
- Completado (escribe 2 letras o Ctrl+Space)
- Tab indenta; atajos +orge / +rule / +ui
- **Live**: al editar, aplica el flow y actualiza preview + estado (debounce)

## Mind agente
Pestaña **Mind agente**: analiza ORGES, reglas, binds, emits y sugiere fragmentos insertables.
Las vistas `(kind UI)` son organismos: Mind las trata como parte del ensamblaje.

## Pulse
Comunicación entre ORGES: tipo, origen, destino, contenido. El runtime firma (Ed25519) y usa nonce anti-replay. Plantilla desde Mind → «Plantilla Pulse».

## UI Alset
`splash`, `gradient` animados, botones `rt-btn`, badges — espacios de UI con bind reactivo.
