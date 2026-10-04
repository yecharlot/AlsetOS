#!/bin/sh
export PATH="/opt/alset/bin:/usr/local/bin:/usr/bin:/bin:$PATH"
export DISPLAY="${DISPLAY:-:0}"
ALSET_ROOT=/opt/alset
DATA="${HOME:-/home/tc}/.alset-desktop"
mkdir -p "$DATA"

# Wait until X is up (max ~60s)
i=0
while [ $i -lt 30 ]; do
  if [ -S /tmp/.X11-unix/X0 ] || xset q >/dev/null 2>&1; then
    break
  fi
  i=$((i+1))
  sleep 2
done

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
# Give bridge a moment
sleep 2

open_browser() {
  if command -v netsurf-gtk >/dev/null 2>&1; then netsurf-gtk "$URL" &
  elif command -v netsurf >/dev/null 2>&1; then netsurf "$URL" &
  elif command -v dillo >/dev/null 2>&1; then dillo "$URL" &
  elif command -v firefox >/dev/null 2>&1; then firefox "$URL" &
  elif command -v midori >/dev/null 2>&1; then midori "$URL" &
  else
    echo "NO_BROWSER $URL" >/tmp/alset-desktop.url
    # last resort: show message in aterm if present
    if command -v aterm >/dev/null 2>&1; then
      aterm -e sh -c "echo AlsetOS Desktop; echo Abre un navegador en $URL; echo; echo Si no hay navegador: tce-load -wi netsurf; sleep 30" &
    fi
  fi
}

open_browser
# retry once later if X was late
sleep 5
open_browser
