#!/usr/bin/env bash
# Dev loop for Alset Desktop bridge + shell (no Tiny Core required).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.26.0}"
go build -o bin/alset-desktop-bridge ./cmd/alset-desktop-bridge
go build -o bin/alsetos ./cmd/alsetos 2>/dev/null || true
echo "Open http://127.0.0.1:7420/"
exec ./bin/alset-desktop-bridge \
  -addr 127.0.0.1:7420 \
  -shell "$ROOT/desktop/shell" \
  -alsetos "$ROOT/bin/alsetos"
