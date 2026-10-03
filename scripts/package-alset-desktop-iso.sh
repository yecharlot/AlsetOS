#!/usr/bin/env bash
# Construye payload listo para USB/ISO (Tiny Core + Alset Desktop + tools web)
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/dist/desktop/alset-os-payload"
WEB="${ALSET_WEB:-$ROOT/../Alset-LISPAI-Runtime/web}"
export GOTOOLCHAIN="${GOTOOLCHAIN:-local}"
export CGO_ENABLED=0

mkdir -p "$OUT/opt/alset/bin" "$OUT/opt/alset/shell" "$OUT/opt/alset/web" "$OUT/opt/bootlocal.d"
cd "$ROOT"
go build -ldflags='-s -w' -o "$OUT/opt/alset/bin/alset-desktop-bridge" ./cmd/alset-desktop-bridge
go build -ldflags='-s -w' -o "$OUT/opt/alset/bin/alsetos" ./cmd/alsetos || true
cp -a desktop/shell/. "$OUT/opt/alset/shell/"
if [[ -d "$WEB" ]]; then
  cp -a "$WEB/." "$OUT/opt/alset/web/"
fi

cat > "$OUT/opt/alset/boot-alset.sh" << 'SH'
#!/bin/sh
ROOT=/opt/alset
"$ROOT/bin/alset-desktop-bridge" \
  -addr 127.0.0.1:7420 \
  -shell "$ROOT/shell" \
  -web "$ROOT/web" \
  -alsetos "$ROOT/bin/alsetos" \
  -data "${HOME:-/tmp}/.alset-desktop" \
  >/tmp/alset-bridge.log 2>&1 &
sleep 1
# abrir browser si existe
URL=http://127.0.0.1:7420/
if command -v firefox >/dev/null 2>&1; then firefox "$URL" &
elif command -v chromium-browser >/dev/null 2>&1; then chromium-browser --app="$URL" &
fi
SH
chmod +x "$OUT/opt/alset/boot-alset.sh" "$OUT/opt/alset/bin/"* 2>/dev/null || true

cat > "$OUT/opt/bootlocal.d/99-alset.sh" << 'SH'
#!/bin/sh
[ -x /opt/alset/boot-alset.sh ] && /opt/alset/boot-alset.sh &
SH
chmod +x "$OUT/opt/bootlocal.d/99-alset.sh"

cat > "$OUT/README-ISO.txt" << 'TXT'
Alset OS Desktop payload
========================
1. Arranca TinyCore desde ISO.
2. Copia este árbol a /opt/alset (y bootlocal.d).
3. /opt/alset/boot-alset.sh
4. Abre http://127.0.0.1:7420/
   - Escritorio + Terminal LispAI
   - Studio: /tools/
   - Editor: /tools/alset-editor/

Remaster ISO (host con xorriso):
  monta TinyCore.iso, copia opt/, enlaza bootlocal, genera AlsetDesktop.iso
TXT

tar -C "$OUT" -czf "$ROOT/dist/desktop/alset-os-desktop-payload.tgz" opt README-ISO.txt
echo "Payload: $ROOT/dist/desktop/alset-os-desktop-payload.tgz"
ls -la "$ROOT/dist/desktop/alset-os-desktop-payload.tgz"
