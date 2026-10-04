#!/usr/bin/env bash
# Build AlsetOS bootable ISO on TinyCore (no loop-mount required; uses xorriso extract).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ISO_IN="${1:-/home/workdir/attachments/TinyCore-current.iso}"
WEB="${ALSET_WEB:-$ROOT/../Alset-LISPAI-Runtime/web}"
WORK="$ROOT/dist/desktop/iso-work"
OUT_ISO="$ROOT/dist/desktop/AlsetOS-Desktop.iso"
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.26.0}"
export CGO_ENABLED=0
export GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}"

[[ -f "$ISO_IN" ]] || { echo "ERROR: ISO base no encontrada: $ISO_IN"; exit 1; }
command -v xorriso >/dev/null || { echo "ERROR: instala xorriso"; exit 1; }

echo "==> Base: $ISO_IN"
rm -rf "$WORK"
mkdir -p "$WORK/iso" "$WORK/core" "$WORK/payload/opt/alset/bin"

# Extract ISO filesystem without mount
echo "==> Extrayendo ISO…"
xorriso -osirrox on -indev "$ISO_IN" -extract / "$WORK/iso" 2>&1 | tail -5

CORE_GZ="$WORK/iso/boot/core.gz"
[[ -f "$CORE_GZ" ]] || CORE_GZ="$(find "$WORK/iso" -name 'core.gz' | head -1)"
[[ -f "$CORE_GZ" ]] || { echo "ERROR: core.gz no encontrado"; find "$WORK/iso" | head -30; exit 1; }
echo "==> core.gz: $CORE_GZ ($(du -h "$CORE_GZ" | cut -f1))"

# Build binaries
cd "$ROOT"
echo "==> Compilando alset-desktop-bridge + alsetos…"
go build -ldflags='-s -w' -o "$WORK/payload/opt/alset/bin/alset-desktop-bridge" ./cmd/alset-desktop-bridge
go build -ldflags='-s -w' -o "$WORK/payload/opt/alset/bin/alsetos" ./cmd/alsetos
chmod +x "$WORK/payload/opt/alset/bin/"*

mkdir -p "$WORK/payload/opt/alset/shell" "$WORK/payload/opt/alset/web"
cp -a "$ROOT/desktop/shell/." "$WORK/payload/opt/alset/shell/"
if [[ -d "$WEB" ]]; then
  echo "==> Web tools (Studio/Editor)…"
  cp -a "$WEB/." "$WORK/payload/opt/alset/web/"
else
  echo "WARN: sin web tools"
fi

cat > "$WORK/payload/opt/alset/boot-alset.sh" << 'SH'
#!/bin/sh
export PATH="/opt/alset/bin:/usr/local/bin:/usr/bin:/bin:$PATH"
ALSET_ROOT=/opt/alset
DATA="${HOME:-/home/tc}/.alset-desktop"
mkdir -p "$DATA"
sleep 2
if [ -x "$ALSET_ROOT/bin/alset-desktop-bridge" ]; then
  "$ALSET_ROOT/bin/alset-desktop-bridge" \
    -addr 0.0.0.0:7420 \
    -shell "$ALSET_ROOT/shell" \
    -web "$ALSET_ROOT/web" \
    -alsetos "$ALSET_ROOT/bin/alsetos" \
    -data "$DATA" \
    >/tmp/alset-bridge.log 2>&1 &
  echo $! >/tmp/alset-bridge.pid
fi
URL="http://127.0.0.1:7420/"
if command -v firefox >/dev/null 2>&1; then firefox "$URL" >/tmp/alset-browser.log 2>&1 &
elif command -v midori >/dev/null 2>&1; then midori "$URL" >/tmp/alset-browser.log 2>&1 &
elif command -v netsurf >/dev/null 2>&1; then netsurf "$URL" >/tmp/alset-browser.log 2>&1 &
else echo "$URL" >/tmp/alset-desktop.url
fi
SH
chmod +x "$WORK/payload/opt/alset/boot-alset.sh"

mkdir -p "$WORK/payload/opt/bootlocal.d"
cat > "$WORK/payload/opt/bootlocal.d/99-alset.sh" << 'SH'
#!/bin/sh
[ -x /opt/alset/boot-alset.sh ] && /opt/alset/boot-alset.sh &
SH
chmod +x "$WORK/payload/opt/bootlocal.d/99-alset.sh"

cat > "$WORK/payload/opt/alset/README.txt" << 'TXT'
AlsetOS Desktop (TinyCore base)
Bridge http://127.0.0.1:7420/
Studio /tools/  Editor /alset-editor/
Manual: /opt/alset/boot-alset.sh
TXT

# Unpack core, inject, repack
echo "==> Descomprimiendo core.gz e inyectando /opt/alset…"
rm -rf "$WORK/core"; mkdir -p "$WORK/core"
( cd "$WORK/core" && gzip -dc "$CORE_GZ" | cpio -idm 2>/dev/null )
cp -a "$WORK/payload/opt" "$WORK/core/"

if [[ -f "$WORK/core/opt/bootlocal.sh" ]]; then
  grep -q '99-alset' "$WORK/core/opt/bootlocal.sh" 2>/dev/null || \
    echo '[ -x /opt/bootlocal.d/99-alset.sh ] && /opt/bootlocal.d/99-alset.sh &' >> "$WORK/core/opt/bootlocal.sh"
else
  mkdir -p "$WORK/core/opt"
  printf '%s\n' '#!/bin/sh' '[ -x /opt/bootlocal.d/99-alset.sh ] && /opt/bootlocal.d/99-alset.sh &' \
    > "$WORK/core/opt/bootlocal.sh"
  chmod +x "$WORK/core/opt/bootlocal.sh"
fi

echo "==> Reempaquetando core.gz (puede tardar)…"
( cd "$WORK/core" && find . | cpio -o -H newc 2>/dev/null | gzip -9 > "$WORK/core-new.gz" )
cp -f "$WORK/core-new.gz" "$CORE_GZ"
ls -lh "$CORE_GZ"

# Cosmetic boot menu
[[ -f "$WORK/iso/boot/isolinux/isolinux.cfg" ]] && \
  sed -i 's/TinyCore/AlsetOS/g; s/tinycore/AlsetOS/g' "$WORK/iso/boot/isolinux/isolinux.cfg" || true
[[ -f "$WORK/iso/boot/isolinux/boot.msg" ]] && \
  echo "AlsetOS Desktop — TinyCore kernel + Alset bridge" > "$WORK/iso/boot/isolinux/boot.msg" || true

# Rebuild ISO (isohybrid-style like original)
echo "==> Generando $OUT_ISO …"
ISOLINUX="boot/isolinux/isolinux.bin"
[[ -f "$WORK/iso/$ISOLINUX" ]] || { echo "ERROR: falta isolinux.bin"; exit 1; }

xorriso -as mkisofs \
  -o "$OUT_ISO" \
  -V "ALSETOS" \
  -J -R -l \
  -b boot/isolinux/isolinux.bin \
  -c boot/isolinux/boot.cat \
  -no-emul-boot \
  -boot-load-size 4 \
  -boot-info-table \
  -isohybrid-mbr /usr/lib/ISOLINUX/isohdpfx.bin \
  "$WORK/iso" 2>&1 | tail -20

# Fallback MBR if isohdpfx missing
if [[ ! -f "$OUT_ISO" ]]; then
  xorriso -as mkisofs \
    -o "$OUT_ISO" \
    -V "ALSETOS" \
    -J -R -l \
    -b boot/isolinux/isolinux.bin \
    -c boot/isolinux/boot.cat \
    -no-emul-boot \
    -boot-load-size 4 \
    -boot-info-table \
    "$WORK/iso" 2>&1 | tail -15
fi

ls -lh "$OUT_ISO"
# Also publish a copy under artifacts root for easy download
cp -f "$OUT_ISO" "$ROOT/../AlsetOS-Desktop.iso" 2>/dev/null || true
cp -f "$OUT_ISO" /home/workdir/artifacts/AlsetOS-Desktop.iso 2>/dev/null || true

# USB how-to
cat > "$ROOT/dist/desktop/USB_BOOT.es.md" << TXT
# AlsetOS Desktop — ISO lista

**Archivo:** \`dist/desktop/AlsetOS-Desktop.iso\`

## USB booteable (Linux)

\`\`\`bash
# Identifica el USB (ej. /dev/sdb) — ¡no uses el disco del sistema!
lsblk
sudo dd if=dist/desktop/AlsetOS-Desktop.iso of=/dev/sdX bs=4M status=progress conv=fsync
\`\`\`

## USB (Windows)

Usa [Rufus](https://rufus.ie) o balenaEtcher en modo DD/imagen ISO.

## Probar en máquina virtual

\`\`\`bash
qemu-system-x86_64 -m 1024 -cdrom dist/desktop/AlsetOS-Desktop.iso -boot d
\`\`\`

## Al arrancar

1. Elige el menú AlsetOS / TinyCore gráfico si aparece.
2. Espera a FLWM / escritorio TC.
3. El bridge Alset debería levantarse solo (\`/opt/bootlocal.d/99-alset.sh\`).
4. Abre el navegador en **http://127.0.0.1:7420/**

Si no abre solo:

\`\`\`sh
/opt/alset/boot-alset.sh
\`\`\`

Log: \`/tmp/alset-bridge.log\`

## Nota

Base: TinyCore (kernel + BusyBox + X).  
Capa Alset: \`/opt/alset\` (bridge, shell, Studio/Editor, alsetos).
TXT

echo
echo "============================================"
echo " ISO: $OUT_ISO"
echo " Guía: $ROOT/dist/desktop/USB_BOOT.es.md"
echo "============================================"
