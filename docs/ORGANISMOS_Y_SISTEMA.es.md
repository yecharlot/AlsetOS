# AlsetOS: todo es organismo + APIs del sistema

## Principio

En Linux “todo es fichero”. En AlsetOS **todo es organismo**: apps, volúmenes, interfaces de red, cuentas, blobs IPFS, el propio escritorio.

## APIs REST (bridge :7420)

| Ruta | Uso |
|------|-----|
| `/v1/organisms` | Lista organismos (`?kind=app\|volume\|network\|ipfs`) |
| `/v1/volumes` `/v1/network` | Volúmenes e interfaces |
| `/v1/ipfs/add\|list\|get` | Store content-addressed local |
| `/v1/auth/login\|me\|accounts` | Cuenta **Master** / master |
| `/v1/audit` | Trazas (Mind observa acciones) |
| `/v1/apps/*` `/v1/deploy` | Apps distintas por nombre |
| `/v1/mcp/tools` `/v1/mcp/call` | Herramientas para agentes externos |

## Apps de fábrica

Organismos, Documentos, Imágenes, Audio/Video (menú), Cuentas, IPFS Store, Archivos, Calculadora.

## Límites honestos

Red/volúmenes se **observan** vía Go (interfaces reales, `/media`/`/mnt`); no configuran el firmware Wi‑Fi del host. IPFS es almacén local con CID. MCP es lite JSON, no el protocolo completo aún.
