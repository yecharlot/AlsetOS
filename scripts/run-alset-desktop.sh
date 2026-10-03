#!/usr/bin/env bash
# Dev: Desktop + Studio + Editor on one port (7420)
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WEB="${ALSET_WEB:-$(cd "$ROOT/../Alset-LISPAI-Runtime/web" 2>/dev/null && pwd || true)}"
export GOTOOLCHAIN="${GOTOOLCHAIN:-local}"
cd "$ROOT"
mkdir -p bin
go build -o bin/alset-desktop-bridge ./cmd/alset-desktop-bridge
go build -o bin/alsetos ./cmd/alsetos 2>/dev/null || true
ARGS=( -addr 127.0.0.1:7420 -shell "$ROOT/desktop/shell" -alsetos "$ROOT/bin/alsetos" )
if [[ -n "${WEB}" && -d "${WEB}" ]]; then
  ARGS+=( -web "$WEB" )
  echo "Tools web: $WEB"
else
  echo "WARN: sin -web (no se encontró Alset-LISPAI-Runtime/web)"
fi
echo "Desktop  http://127.0.0.1:7420/"
echo "Studio   http://127.0.0.1:7420/tools/  (index Studio)"
echo "Editor   http://127.0.0.1:7420/tools/alset-editor/"
exec ./bin/alset-desktop-bridge "${ARGS[@]}"
