#!/usr/bin/env bash
set -euo pipefail
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.26.0}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/dist"
mkdir -p "$OUT"
TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "linux/arm"
  "darwin/arm64"
  "windows/amd64"
)
CMDS=(alsetos alset-node alset-kernel alset-p2p)
for t in "${TARGETS[@]}"; do
  GOOS="${t%/*}"; GOARCH="${t#*/}"
  for c in "${CMDS[@]}"; do
    name="alset-${c#alset}"
    [[ "$c" == alsetos ]] && name=alsetos
    out="$OUT/${name}-${GOOS}-${GOARCH}"
    [[ "$GOOS" == windows ]] && out="${out}.exe"
    echo "building $out"
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags="-s -w" -o "$out" "./cmd/$c"
  done
done
echo "OK → $OUT"
ls -la "$OUT"
