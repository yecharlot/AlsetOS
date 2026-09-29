#!/usr/bin/env bash
# Build a bootable payload folder + optional ISO if xorriso/genisoimage exists.
# Tiny Core remains the kernel/base; Alset packages live under /opt/alset.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ISO_IN="${1:-}"
OUT="${2:-$ROOT/dist/desktop/alset-desktop-rootfs}"
PKG="$ROOT/dist/desktop/package"

if [[ ! -d "$PKG/bin" ]]; then
  echo "Build package first: binaries in dist/desktop/package"
  exit 1
fi

rm -rf "$OUT"
mkdir -p "$OUT/opt/alset" "$OUT/opt/bootlocal.d" "$OUT/tce/optional"
cp -a "$PKG/." "$OUT/opt/alset/"
cat > "$OUT/opt/bootlocal.d/99-alset-desktop.sh" << 'SH'
#!/bin/sh
# Runs late in TC bootlocal if linked from /opt/bootlocal.sh
ALSET_ROOT=/opt/alset
export ALSET_ROOT
[ -x "$ALSET_ROOT/boot/alset-desktop.sh" ] && "$ALSET_ROOT/boot/alset-desktop.sh" &
SH
chmod +x "$OUT/opt/bootlocal.d/99-alset-desktop.sh" "$OUT/opt/alset/boot/alset-desktop.sh" || true

# seed desktop package as .alset
mkdir -p "$OUT/opt/alset/pkg"
cat > "$OUT/opt/alset/pkg/desktop-shell.alset.json" << JSON
{
  "format": "alset-pkg/v1",
  "name": "desktop-shell",
  "version": "0.1.0",
  "kind": "shell",
  "description": "Alset Desktop Shell + bridge entry",
  "entrypoint": "/opt/alset/boot/alset-desktop.sh"
}
JSON

echo "Rootfs Alset listo: $OUT"
echo "Contiene /opt/alset (bridge, shell, alsetos, .alset pkg)"
echo
echo "Para remaster ISO completo en una máquina con xorriso:"
echo "  1) Monta TinyCore-current.iso"
echo "  2) Copia $OUT/opt/alset dentro del árbol de arranque"
echo "  3) Añade a bootlocal.sh: /opt/bootlocal.d/99-alset-desktop.sh"
echo "  4) xorriso -as mkisofs ... -o AlsetDesktop.iso"
echo
if [[ -n "$ISO_IN" && -f "$ISO_IN" ]]; then
  echo "ISO de entrada detectada: $ISO_IN"
  echo "Copia manual del payload a un USB de datos junto al ISO live:"
  echo "  mkdir -p /mnt/usb/opt && cp -a $OUT/opt/alset /mnt/usb/opt/"
  echo "  En TC: monta USB y sudo cp -a /mnt/usb/opt/alset /opt/"
fi

# tarball payload for USB
tar -C "$OUT" -czf "$ROOT/dist/desktop/alset-desktop-payload.tgz" opt
echo "Payload: $ROOT/dist/desktop/alset-desktop-payload.tgz"
