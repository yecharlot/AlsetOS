#!/bin/sh
# AlsetOS Desktop — bridge + kiosk (sin escritorio TinyCore visible)
export PATH="/opt/alset/bin:/usr/local/bin:/usr/bin:/bin:$PATH"
export DISPLAY="${DISPLAY:-:0}"
export HOME="${HOME:-/home/tc}"
ALSET_ROOT=/opt/alset
DATA="$HOME/.alset-desktop"
URL="http://127.0.0.1:7420/"
mkdir -p "$DATA" /tmp

log() { echo "$(date -u +%H:%M:%S) $*" >> /tmp/alset-boot.log; }
log "boot-alset start DISPLAY=$DISPLAY"

# Wait for X (up to ~90s)
i=0
while [ $i -lt 45 ]; do
  if [ -S /tmp/.X11-unix/X0 ] 2>/dev/null; then log "X socket ok"; break; fi
  if xset q >/dev/null 2>&1; then log "xset ok"; break; fi
  i=$((i + 1))
  sleep 2
done

# Start bridge if not running
if ! wget -q -O /dev/null --timeout=2 "$URL" 2>/dev/null \
  && ! curl -sf --max-time 2 "$URL" >/dev/null 2>&1; then
  if [ -x "$ALSET_ROOT/bin/alset-desktop-bridge" ]; then
    log "starting bridge"
    "$ALSET_ROOT/bin/alset-desktop-bridge" \
      -addr 0.0.0.0:7420 \
      -shell "$ALSET_ROOT/shell" \
      -web "$ALSET_ROOT/web" \
      -alsetos "$ALSET_ROOT/bin/alsetos" \
      -data "$DATA" \
      >>/tmp/alset-bridge.log 2>&1 &
    echo $! >/tmp/alset-bridge.pid
  else
    log "ERROR: bridge binary missing"
  fi
fi

# Wait until bridge answers
j=0
while [ $j -lt 30 ]; do
  if wget -q -O /dev/null --timeout=2 "$URL" 2>/dev/null; then log "bridge up"; break; fi
  if curl -sf --max-time 2 "$URL" >/dev/null 2>&1; then log "bridge up curl"; break; fi
  j=$((j + 1))
  sleep 1
done

# Hide typical TC desktop noise if present
killall wbar 2>/dev/null || true

open_kiosk() {
  log "open_kiosk $URL"
  # Prefer browsers that can feel fullscreen
  if command -v chromium >/dev/null 2>&1; then
    chromium --kiosk --noerrdialogs --disable-session-crashed-bubble \
      --check-for-update-interval=31536000 --app="$URL" >>/tmp/alset-browser.log 2>&1 &
  elif command -v chromium-browser >/dev/null 2>&1; then
    chromium-browser --kiosk --app="$URL" >>/tmp/alset-browser.log 2>&1 &
  elif command -v google-chrome >/dev/null 2>&1; then
    google-chrome --kiosk --app="$URL" >>/tmp/alset-browser.log 2>&1 &
  elif command -v firefox >/dev/null 2>&1; then
    firefox -kiosk "$URL" >>/tmp/alset-browser.log 2>&1 &
  elif command -v netsurf-gtk >/dev/null 2>&1; then
    netsurf-gtk "$URL" >>/tmp/alset-browser.log 2>&1 &
  elif command -v netsurf >/dev/null 2>&1; then
    netsurf "$URL" >>/tmp/alset-browser.log 2>&1 &
  elif command -v midori >/dev/null 2>&1; then
    midori -e Fullscreen -a "$URL" >>/tmp/alset-browser.log 2>&1 &
  elif command -v dillo >/dev/null 2>&1; then
    dillo "$URL" >>/tmp/alset-browser.log 2>&1 &
  else
    log "NO_BROWSER"
    echo "$URL" >/tmp/alset-desktop.url
    if command -v aterm >/dev/null 2>&1; then
      aterm -geometry 100x30 -e sh -c "echo AlsetOS; echo Abre navegador en $URL; sleep 60" &
    fi
  fi
}

open_kiosk
# Reintentos si X o el browser fallan al inicio
sleep 4
open_kiosk
sleep 8
open_kiosk
log "boot-alset done"
